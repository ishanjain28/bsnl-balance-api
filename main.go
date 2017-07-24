package main

import (
	"github.com/gorilla/mux"
	"net/http"
	"os"
	"log"
	"time"
	"strconv"
	"encoding/json"
	"io/ioutil"
	"fmt"
	"github.com/getlantern/errors"
	"crypto/tls"
	"strings"
	"net/url"
	"bytes"
)

var (
	POC *PostpaidCircles
	PRC *PrepaidCircles
)

type PostpaidCircles struct {
	ROWSET struct {
		ROW []struct {
			CIRCLEID     string `json:"CIRCLE_ID"`
			CIRCLENAME   string `json:"CIRCLE_NAME"`
			ZONEID       string `json:"ZONE_ID"`
			ZONENAME     string `json:"ZONE_NAME"`
			CMZONECODE   string `json:"CM_ZONE_CODE"`
			CMCIRCLECODE string `json:"CM_CIRCLE_CODE"`
			CIRCLECODE   string `json:"CIRCLE_CODE"`
			ZONECODE     string `json:"ZONE_CODE"`
		} `json:"ROW"`
	} `json:"ROWSET"`
}

type PrepaidCircles struct {
	ROWSET struct {
		ROW []struct {
			CIRCLEID   int `json:"CIRCLE_ID"`
			CIRCLENAME string `json:"CIRCLE_NAME"`
			ZONEID     int `json:"ZONE_ID"`
			ZONENAME   string `json:"ZONE_NAME"`
			CIRCLECODE string `json:"CIRCLE_CODE"`
			ZONECODE   string `json:"ZONE_CODE"`
		} `json:"ROW"`
	} `json:"ROWSET"`
}

type BsnlRequest struct {
	USERID             string `json:"USERID"`
	PHONENO            string `json:"PHONE_NO"`
	PREPAIDNO          string `json:"PREPAIDNO"`
	EMAILID            string `json:"EMAILID"`
	CONTACTNO          string `json:"CONTACTNO"`
	SHORTNAME          string `json:"SHORTNAME"`
	SVCTYPE            string `json:"SVC_TYPE"`
	SSACODE            string `json:"SSA_CODE"`
	ZONECODE           string `json:"ZONE_CODE"`
	CIRCLEID           int `json:"CIRCLE_ID"`
	CIRCLECODE         string `json:"CIRCLE_CODE"`
	ACCOUNTNO          string `json:"ACCOUNT_NO"`
	DENOMINATION       string `json:"DENOMINATION"`
	TOTALAMOUNT        string `json:"TOTAL_AMOUNT"`
	INVOICENO          string `json:"INVOICE_NO"`
	INVOICEDATE        string `json:"INVOICE_DATE"`
	DUEDATE            string `json:"DUE_DATE"`
	AGENCY             string `json:"AGENCY"`
	VOUCHERCATEGORY    string `json:"VOUCHER_CATEGORY"`
	VOUCHERSUBCATEGORY string `json:"VOUCHER_SUBCATEGORY"`
}

type BsnlSucess struct {
	STATUS  string `json:"STATUS"`
	REMARKS string `json:"REMARKS"`
	BALANCE string `json:"BALANCE"`
}

type BsnlFailure struct {
	STATUS  string `json:"STATUS"`
	REMARKS string `json:"REMARKS"`
}

func init() {
	var err error
	POC, err = fetchPostpaidCircles()
	if err != nil {
		log.Println("Error in fetching postpaid circles", err.Error())
	}

	PRC, err = fetchPrepaidCircles()
	if err != nil {
		log.Println("Error in fetching prepaid circles", err.Error())
	}
}

func main() {

	PORT := os.Getenv("PORT")
	if PORT == "" {
		log.Fatalln("$PORT not set")
	}

	router := mux.NewRouter()

	router.HandleFunc("/balance/{phone}/{circle-code}", fetchBalance)

	router.HandleFunc("/", homeHandler)

	fmt.Println("Starting Server on", PORT)
	http.ListenAndServe(":"+PORT, router)
}

func createBSNLRequest(phoneno, cCode string) (*BsnlRequest, error) {

	if phoneno == "" || cCode == "" {
		return nil, errors.New("Invalid Input")
	}

	breq := &BsnlRequest{}

	breq.USERID = "0"
	breq.EMAILID = ""
	breq.CONTACTNO = ""
	breq.SHORTNAME = ""
	breq.SVCTYPE = "PPGSM"
	breq.SSACODE = "NA"
	breq.ACCOUNTNO = "NA"
	breq.DENOMINATION = "0"
	breq.TOTALAMOUNT = "0"
	breq.INVOICEDATE = "NA"
	breq.INVOICENO = "NA"
	breq.DUEDATE = "NA"
	breq.AGENCY = "PORTAL"
	breq.VOUCHERCATEGORY = "T"
	breq.VOUCHERSUBCATEGORY = "NA"
	breq.PHONENO = phoneno
	breq.PREPAIDNO = phoneno

	for _, v := range PRC.ROWSET.ROW {
		if strings.ToLower(v.CIRCLECODE) == strings.ToLower(cCode) {
			breq.ZONECODE = v.ZONECODE
			breq.CIRCLECODE = v.CIRCLECODE
			breq.CIRCLEID = v.CIRCLEID
		}
	}

	return breq, nil
}

func fetchBalance(w http.ResponseWriter, r *http.Request) {

	v := mux.Vars(r)
	phone := v["phone"]
	cCode := v["circle-code"]

	breq, err := createBSNLRequest(phone, cCode)
	if err != nil {
		log.Println("Error in creating request", err.Error())
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	t, err := json.Marshal(breq)
	if err != nil {
		log.Println("Error in marshalling", err.Error())
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	reqData := []byte(url.PathEscape("postData=" + string(t)))

	req, err := http.NewRequest("POST", "https://portal2.bsnl.in/myportal/validatepprequest.do", bytes.NewBuffer(reqData))

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Origin", "https://portal2.bsnl.in")
	req.Header.Set("Referer", "https://portal2.bsnl.in/myportal/workspace.do")

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true },
	}
	client := &http.Client{Transport: tr}

	resp, err := client.Do(req)

	if err != nil {
		log.Fatalln(err.Error())
	}

	rData, _ := ioutil.ReadAll(resp.Body)

	if strings.Index(string(rData), "SUCCESS") {
		bs := &BsnlSucess{}
		err = json.Unmarshal(rData, bs)

		if err != nil {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			log.Println(err.Error())
			return
		}

		bs.
	} else {
		bf := &BsnlFailure{}
		err = json.Unmarshal(rData, bf)

		if err != nil {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			log.Println(err.Error())
			return
		}
	}

}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "https://market.mashape.com/bsnl-balance", http.StatusTemporaryRedirect)
}

func fetchPrepaidCircles() (*PrepaidCircles, error) {

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}

	resp, err := client.Get("http://portal2.bsnl.in/myportal/JSON/circles_prepaid.json?" + strconv.FormatInt(time.Now().UnixNano(), 10))
	if err != nil {
		return nil, err
	}

	p := &PrepaidCircles{}

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(b, p)
	if err != nil {
		return nil, err
	}

	return p, nil

}

func fetchPostpaidCircles() (*PostpaidCircles, error) {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}

	resp, err := client.Get("http://portal2.bsnl.in/myportal/JSON/circles_postpaid.json?" + strconv.FormatInt(time.Now().UnixNano(), 10))
	if err != nil {
		return nil, err
	}

	p := &PostpaidCircles{}

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(b, p)
	if err != nil {
		return nil, err
	}

	return p, nil
}
