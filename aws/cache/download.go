// Without requiring authentication, download pricing list in bulk
// via https. For EC2 prices, the volume of the download is about
// 1GB. The json returned by the bulk API is one object with a very
// large map inside it, so the code has to load the entire Json into
// memory before parsing it. If this turns out to be a problem, will
// have to call GetAttributeValues on the endpoint with pagination,
// but that does require authentication.
package cache

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
)

func getPricesInBulk(db *sql.DB, url string, filename string) error {
	var decoded map[string]interface{}
	if url != "" {
		c := &http.Client{Timeout: 200 * time.Second}
		//response, err := c.Get("https://pricing.us-east-1.amazonaws.com/offers/v1.0/aws/AmazonEC2/current/index.json")
		response, err := c.Get(url)

		if err != nil {
			return err
		}
		defer response.Body.Close()
		err = json.NewDecoder(response.Body).Decode(&decoded)
		if err != nil {
			return err
		}
	} else if filename != "" {
		file, _ := os.Open(filename)
		defer file.Close()
		decoder := json.NewDecoder(file)
		for decoder.More() {
			fmt.Println("one item in decoder")
			decoder.Decode(&decoded)
		}
	} else {
		return fmt.Errorf("need url or filename for getting prices")
	}
	publicationDate, _ := decoded["publicationDate"].(string)
	t, err := time.Parse("2006-01-02T15:04:05Z", publicationDate)
	if err != nil {
		return err
	}
	var counter uint32
	products, _ := decoded["products"].(map[string]interface{})
	fmt.Printf("publication date: %+v has %d products\n", t, len(products))
	for sku, product := range products {
		counter++
		if counter > 10 {
			break
		}
		if counter%1000 == 0 {
			fmt.Printf("counter: %d\n", counter)
		}
		p, _ := product.(map[string]interface{})
		pf, _ := p["productFamily"].(string)
		attributes, hasAttr := p["attributes"].(map[string]interface{})
		// Maybe it would be better to get the instance types by
		// running describe-instance-types.
		if hasAttr {
			// Not covering "Dedicated Host" for now.
			if pf == "Compute Instance" || pf == "Compute Instance (bare metal)" {
				fmt.Printf("sku %s is for an instance with attributes %+v\n",
				sku, attributes)
				/*
					if err := insertInstanceType(db, sku, attributes); err != nil {
						return err
					}
					if err := insertInstanceSku(db, sku, attributes); err != nil {
						return err
					}
				*/
			} else if pf == "Storage" {
				fmt.Printf("sku %s storage: %v\n", sku, attributes)
				/*
				if err := insertVolumeType(db, attributes); err != nil {
					return err
				}
				*/

			} else {
				fmt.Printf(
					"sku %s: missing handler for product family %s\n", sku, pf)
			}
		} else {
			fmt.Printf("no attr on %+v\n", p)
		}
		// populate pricing tiers
	}
	return nil
}

func insertVolumeType(db *sql.DB, attributes map[string]interface{}) error {
	vtype, _ := attributes["volumeType"].(string) // is this right?
	media, _ := attributes["storageMedia"].(string)

	rawSize, _ := attributes["maxVolumeSize"].(string)
	var size uint32
	var sizeGb uint32
	res, err := fmt.Sscanf(rawSize, "%d TiB", &size)
	if err != nil && res == 1 {
		sizeGb = size * 1000 // or 1024?
	} else {
		return fmt.Errorf("failed to parse maxVolumeSize %s: %v %d\n", rawSize, err, res)
	}
	rawIops, _ := attributes["maxIopsvolume"].(string)
	var iops uint32
	tmp, err := strconv.Atoi(rawIops)
	iops = uint32(tmp)
	if err != nil {
		var tmp int
		res, err = fmt.Sscanf(rawIops, "%d - based on %d MiB I/O size", &iops, &tmp)
		if err != nil || res != 1 {
			res, err = fmt.Sscanf(rawIops, "%d - %d", &tmp, &iops)
			if err != nil || res != 1 {
				return fmt.Errorf("failed to parse maxIopsvolume %s\n", rawIops)
			}
		}
		log.Printf("parsed %d iops out of %s\n", iops, rawIops)
	}
	rawThroughput, _ := attributes["maxThroughputvolume"].(string)
	var throughput uint32
	res, err = fmt.Sscanf(rawThroughput, "%d MiB/s", &throughput)
	if err != nil || res != 1 {
		return fmt.Errorf("failed to parse throughput %s\n", rawThroughput)
	}
	insert := `REPLACE INTO VolumeTypes (VolumeType, StorageMedia,
	MaxVolumeSize, MaxIOPS, MaxThroughput) VALUES (?, ?, ?, ?, ?);`
	stmt, err := db.Prepare(insert)
	_, err = stmt.Exec(vtype, media, sizeGb, iops, throughput)
	if err != nil {
		return err
	}
	region, err := getRegion(db, attributes)
	if err != nil {
		return err
	}
	if region != "" {
		regionInsert := `REPLACE INTO VolumeTypeByRegion (
			VolumeType, Region) VALUES (?, ?);`
		stmt, err = db.Prepare(regionInsert)
		_, err = stmt.Exec(vtype, region)
		if err != nil {
			return err
		}
	} else {
		fmt.Printf("missing handler for location type %s for volume type %s\n",
			attributes["locationType"], vtype)
	}
	return nil
}

func insertInstanceSku(db *sql.DB, sku string, attributes map[string]interface{}) error {
	var itype string
	itype, ok := attributes["instanceType"].(string)
	if !ok {
		return fmt.Errorf("attribute map %v for sku %s missing instance type", attributes, sku)
	}
	region, err := getRegion(db, attributes)
	if err != nil {
		return err
	}
	// regexp for usage: shortlocation-Usage:instance type
	re := regexp.MustCompile(`([A-Z1-9-]+-)?([^:]+):.*`)
	usagePieces := re.FindStringSubmatch(attributes["usagetype"].(string))
	usage := "BoxUsage"
	if len(usagePieces) == 3 {
		usage = usagePieces[2]
	} else if len(usagePieces) == 2 {
		usage = usagePieces[1]
	} else {
		return fmt.Errorf("Failed to parse usage type %s\n", attributes["usagetype"].(string))
	}
	operation := "RunInstances"
	op, _ := attributes["operation"].(string)
	if op == "Hourly" {
		operation = "Hourly"
	} else {
		// regexp for operation: Hourly or RunInstances[:[0-9]4]?.
		parts := strings.Split(op, ":")
		if len(parts) != 2 {
			return fmt.Errorf("operation format not recognised: %s\n", op)
		}
		if strings.HasPrefix(parts[1], "FFP") {
			log.Printf("RunInstances:FFP codes only exist in GovCloud, not supported.\n")
			return nil
		}
		operation = parts[1]
	}
	insert := `REPLACE INTO Sku (Sku, ProductType, Region, Usage, Operation)
	VALUES (?, ?, ?, ?, ?);`
	stmt, err := db.Prepare(insert)
	_, err = stmt.Exec(sku, itype, region, usage, operation)
	if err != nil {
		return err
	}
	return nil
}

func insertInstanceType(db *sql.DB, sku string, attributes map[string]interface{}) error {
	var itype string
	var ifamily string
	itype, ok := attributes["instanceType"].(string)
	if !ok {
		return fmt.Errorf("attribute map %v for sku %s missing instance type", attributes, sku)
	}
	ifamily, ok = attributes["instanceFamily"].(string)
	if !ok {
		return fmt.Errorf("attribute map %v missing instance family",
			attributes)
	}
	var vcpu uint32
	fmt.Sscanf(attributes["vcpu"].(string), "%d", &vcpu)
	var memory uint32
	fmt.Sscanf(attributes["memory"].(string), "%d GiB", &memory)
	var gpu uint32
	if gpuspec, ok := attributes["gpu"].(string); ok {
		fmt.Sscanf(gpuspec, "%d", &gpu)
	}
	storageType, sizeGb, err := parseStorageSpec(attributes["storage"].(string))
	if err != nil {
		return err
	}
	insert := `INSERT INTO InstanceTypes (InstanceType, InstanceFamily,
	CPU, Memory, GPU, StorageType, StorageAmount)
	VALUES (?, ?, ?, ?, ?, ?, ?);`
	stmt, err := db.Prepare(insert)
	_, err = stmt.Exec(itype, ifamily, vcpu, memory, gpu, storageType, sizeGb)
	if err != nil {
		return err
	}
	region, err := getRegion(db, attributes)
	if err != nil {
		return err
	}
	if region != "" {
		regionInsert := `INSERT INTO InstanceTypeByRegion (
			InstanceType, Region) VALUES (?, ?);`
		stmt, err = db.Prepare(regionInsert)
		_, err = stmt.Exec(attributes["instanceType"], region)
		if err != nil {
			return err
		}
	} else {
		fmt.Printf("missing handler for location type %s for instance type %s\n",
			attributes["locationType"], attributes["instanceType"])
	}
	return nil
}

func getRegion(db *sql.DB, attributes map[string]interface{}) (string, error) {
	if attributes["locationType"].(string) == "AWS Region" {
		regionId, err := RegionByDisplayName(
			db, attributes["location"].(string))
		if err != nil {
			return "", err
		}
		if regionId == "" {
			log.Printf("region %s is not supported.\n",
				attributes["location"].(string))
			return "", nil
		}
		return regionId, nil
	}
	log.Printf("unsupported location type: %s\n",
		attributes["locationType"].(string))
	return "", nil
}

func parseStorageSpec(spec string) (string, uint32, error) {
	if spec == "EBS only" {
		return spec, 0, nil
	}
	var times uint32
	var sizeGb uint32
	var diskTech string
	// Maybe this should use a regexp.
	res, _ := fmt.Sscanf(spec, "%d x %d %s", &times, &sizeGb, &diskTech)
	if res != 3 {
		res, _ := fmt.Sscanf(spec, "%d GB %s", &sizeGb, &diskTech)
		if res != 2 {
			return "", 0, fmt.Errorf("failed to parse storage spec %s",
				spec)
		}
		times = 1
	}
	if diskTech == "NvMe" {
		diskTech = "NvMe SSD"
	}
	return diskTech, times * sizeGb, nil
}
