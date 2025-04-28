package main

import (
	"flag"
	"fmt"
	"log"
)

func main() {
	// Parse command-line arguments
	identifier := flag.String("identifier", "", "Strato identifier")
	password := flag.String("password", "", "Strato password")
	order := flag.String("order", "", "Package order number to update")
	domain := flag.String("domain", "", "(Sub-)Domain to manage")
	command := flag.String("command", "", "Command to execute: add, remove, or list")
	recordType := flag.String("type", "TXT", "Type of DNS record (default: TXT)")
	recordPrefix := flag.String("prefix", "", "Prefix for the DNS record")
	recordValue := flag.String("value", "", "Value for the DNS record")
	flag.Parse()

	if *identifier == "" || *password == "" || *order == "" || *domain == "" || *command == "" {
		log.Fatal("All flags --identifier, --password, --order, --domain, and --command are required")
	}

	// Initialize the Strato client
	client, err := NewStratoClient(*identifier, *password, *order, *domain)
	if err != nil {
		log.Fatalf("Failed to create Strato client: %v", err)
	}

	// Execute command
	switch *command {
	case "list":
		config, err := client.GetDNSConfiguration()
		if err != nil {
			log.Fatalf("Failed to fetch DNS records: %v", err)
		}
		printConfig(config)
		return

	case "add":
		if *recordType == "" {
			log.Fatal("--type is required for add command")
		}
		if *recordPrefix == "" {
			log.Fatal("--prefix is required for add command")
		}
		if *recordValue == "" {
			log.Fatal("--value is required for add command")
		}
		providedRecord := DNSRecord{
			Type:   *recordType,
			Prefix: *recordPrefix,
			Value:  *recordValue,
		}
		config, err := client.GetDNSConfiguration()
		if err != nil {
			log.Fatalf("Failed to fetch initial configuration: %v", err)
			return
		}
		fmt.Println("DNS configuration before update:")
		printConfig(config)

		if contains(config.Records, providedRecord) {
			log.Printf("Record already exists: Type: '%s', Prefix: '%s', Value: '%s'\n", providedRecord.Type, providedRecord.Prefix, providedRecord.Value)
			return
		}

		config.Records = append(config.Records, providedRecord)
		if err := client.SetDNSConfiguration(config); err != nil {
			log.Fatalf("Failed to update DNS records: %v", err)
		}
		config, err = client.GetDNSConfiguration()
		if err != nil {
			log.Fatalf("Failed to fetch updated configuration: %v", err)
		}
		printConfig(config)
		if !contains(config.Records, providedRecord) {
			log.Fatalf("Failed to add new record")
			return
		}
		fmt.Println("New record added successfully")
		return
	case "remove":
		if *recordType == "" {
			log.Fatal("--type is required for add command")
		}
		if *recordPrefix == "" {
			log.Fatal("--prefix is required for add command")
		}
		if *recordValue == "" {
			log.Fatal("--value is required for add command")
		}
		providedRecord := DNSRecord{
			Type:   *recordType,
			Prefix: *recordPrefix,
			Value:  *recordValue,
		}
		config, err := client.GetDNSConfiguration()
		if err != nil {
			log.Fatalf("Failed to fetch initial configuration: %v", err)
		}
		fmt.Println("DNS configuration before update:")
		printConfig(config)

		var updatedRecords []DNSRecord
		for _, record := range config.Records {
			if record.Type != providedRecord.Type || record.Prefix != providedRecord.Prefix || record.Value != providedRecord.Value {
				updatedRecords = append(updatedRecords, record)
			}
		}
		if len(updatedRecords) == len(config.Records) {
			log.Printf("Record not found: Type: '%s', Prefix: '%s', Value: '%s'\n", providedRecord.Type, providedRecord.Prefix, providedRecord.Value)
			return
		}
		config.Records = updatedRecords

		if err := client.SetDNSConfiguration(config); err != nil {
			log.Fatalf("Failed to update DNS configuration: %v", err)
		}
		config, err = client.GetDNSConfiguration()
		if err != nil {
			log.Fatalf("Failed to fetch DNS configuration: %v", err)
		}
		fmt.Println("DNS configuration after update:")
		printConfig(config)
		if contains(config.Records, providedRecord) {
			log.Fatalf("Failed to remove record")
			return
		}
		fmt.Println("Record successfully removed")
		return
	default:
		log.Fatalf("Invalid command: %s. Use add, remove, or list", *command)
	}
}

func printConfig(config DNSConfig) {
	fmt.Println("DMARC Type:", config.DMARCType)
	fmt.Println("SPF Type:", config.SPFType)
	fmt.Println("DNS records:")
	for _, record := range config.Records {
		fmt.Printf("Type: '%s', Prefix: '%s', Value: '%s'\n", record.Type, record.Prefix, record.Value)
	}
}

func contains(records []DNSRecord, record DNSRecord) bool {
	for _, entry := range records {
		if entry.Type == record.Type && entry.Prefix == record.Prefix && entry.Value == record.Value {
			return true
		}
	}
	return false
}
