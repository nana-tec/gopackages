package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	linkvaluer "github.com/nana-tec/gopackages/LinkValuer"
)

func main() {
	email := os.Getenv("LINKVALUER_EMAIL")
	pass := os.Getenv("LINKVALUER_PASSWORD")
	if email == "" || pass == "" {
		log.Fatal("set LINKVALUER_EMAIL and LINKVALUER_PASSWORD env vars before running")
	}

	dl_ := ""

	cfg := &linkvaluer.Config{
		Credentials: linkvaluer.Credentials{Email: email, Password: pass},
		Debug:       true,
		TokenTTL:    6 * time.Hour,
	}
	c, err := linkvaluer.NewClient(cfg)
	if err != nil {
		log.Fatalf("new client: %v", err)
	}

	if err := c.Login(); err != nil {
		log.Fatalf("login: %v", err)
	}
	fmt.Println("Login successful. Access token present:", c.IsTokenValid())

	// Create valuation (sample data)
	createReq := &linkvaluer.CreateRequest{
		CustomerName:       "Test User",
		CustomerPhone:      "0712345678",
		RegistrationNumber: "KAA000A",
		PolicyNumber:       "POL123",
		CustomerEmail:      "test@example.com",
		InsuranceCompany:   "Ibima",
		CallBackURL:        "https://example.com/callback",
		PartnerReference:   "PARTNER123",
	}
	resp, err := c.CreateValuation(createReq)
	if err != nil {
		log.Printf("create valuation error: %v", err)
	} else {
		fmt.Println("Create valuation response message:", resp.Message)
	}

	// View assessments
	assessments, err := c.ViewAssessments()
	if err != nil {
		log.Printf("view assessments error: %v", err)
	} else {
		fmt.Printf("Assessments: %d items (page %d/%d)\n", len(assessments.Data), assessments.Pagination.CurrentPage, assessments.Pagination.LastPage)
		// Pretty-print a small preview by re-encoding the typed payload
		b, _ := json.Marshal(assessments)
		var pretty bytes.Buffer
		if err := json.Indent(&pretty, b, "", "  "); err == nil {
			fmt.Println("Assessments preview (truncated):")
			//pb := pretty.Bytes()
			////if len(pb) > 1000 {
			////	pb = pb[:1000]
			//}
			fmt.Println(assessments)
		}

		// Print a concise summary for first few
		for i := 0; i < len(assessments.Data) && i < 3; i++ {
			it := assessments.Data[i]
			dl := ""
			if it.DownloadURL != nil {
				dl = *it.DownloadURL
				if dl != "" {
					dl_ = it.BookingNo
				}
			}
			fmt.Printf("- %s | %s | %s | %s\n", it.BookingNo, it.RegNo, it.Status, dl)
		}
	}

	// Download a report if you have a booking number
	booking := dl_
	if booking != "" {
		bytes, ct, err := c.DownloadReport(booking)
		if err != nil {
			log.Printf("download report error: %v", err)
		} else {
			fmt.Println("Report content-type:", ct)
			if err := os.WriteFile("report.pdf", bytes, 0644); err != nil {
				log.Printf("write file error: %v", err)
			} else {
				fmt.Println("Saved report to report.pdf")
			}
		}
	}
}
