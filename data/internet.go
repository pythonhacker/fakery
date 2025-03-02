package data

import (
	"log"
)

// Internet related data
var (
	netDataLoader    JSONDataLoader
	WeightedTLD      *WeightedArray
	FakeEmailDomains []string
)

func init() {
	var err error
	if err = netDataLoader.Load(genericLocale, "internet.json"); err != nil {
		log.Fatalf("error loading internet data - %v", err)
	}
	WeightedTLD, err = netDataLoader.GetWeightedArray("common_tlds_weighted", ":")
	if err != nil {
		log.Fatalf("error loading tld data - %v", err)
	}

	FakeEmailDomains = netDataLoader.Get("fake_email_domains")
}
