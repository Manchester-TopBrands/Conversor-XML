package main

import (
	"encoding/xml"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"text/template"

	"github.com/tealeg/xlsx"

	_ "embed"
)

//go:embed page1.html
var index string

type Prod struct {
	Cprod       string  `xml:"cProd"`
	CEAN        string  `xml:"cEAN"`
	NCM         string  `xml:"NCM,omitempty"`
	DescProduto string  `xml:"xProd,omitempty"`
	Quantidade  float64 `xml:"qCom,omitempty"`
	ValorUni    float64 `xml:"vUnCom"`
}

type Det struct {
	Prod Prod `xml:"prod"`
}

type InfNFe struct {
	Det []Det `xml:"det"`
}

type NFe struct {
	InfNFe InfNFe `xml:"infNFe"`
}

type DataFormat struct {
	NFe NFe `xml:"NFe"`
}

var tmpl *template.Template

var createConfig bool

func main() {

	flag.BoolVar(&createConfig, "c", false, "create config.yaml file")
	flag.Parse()

	if createConfig {
		createConfigFile()
		return
	}

	var err error
	tmpl, err = template.New("index").Parse(index)
	if err != nil {
		log.Fatal(err)
	}

	log.Print("loading config file")
	if err := loadConfig(); err != nil {
		log.Fatal(err)
	}

	log.Printf("starting server '%s' at port: %s", config.API.Host, config.API.Port)
	http.HandleFunc("/", redirect)
	http.HandleFunc("/index.html", homeHandler)
	http.HandleFunc("/xml", apiHandler)
	log.Fatal(http.ListenAndServe(":"+config.API.Port, nil))

}
func redirect(w http.ResponseWriter, req *http.Request) {
	// remove/add not default ports from req.Host
	target := "http://" + req.Host + "/index.html"
	// log.Println(target)
	if len(req.URL.RawQuery) > 0 {
		target += "?" + req.URL.RawQuery
	}
	// log.Printf("redirect to: %s", target)
	http.Redirect(w, req, target,
		// see comments below and consider the codes 308, 302, or 301
		http.StatusTemporaryRedirect)
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	tmpl.Execute(w, struct {
		Host string
		Port string
	}{config.API.Host, config.API.Port})
}

// func handler2(w http.ResponseWriter, r *http.Request) {
// 	file, header, err := r.FormFile("file")
// 	if err != nil {
// 		w.WriteHeader(http.StatusBadRequest)
// 		return
// 	}
// 	b, err := ioutil.ReadAll(file)
// 	if err != nil {
// 		w.WriteHeader(http.StatusBadRequest)
// 		return
// 	}

// 	ioutil.WriteFile(header.Filename, b, 0766)

// 	w.Header().Set("Access-Control-Allow-Origin", "*")
// 	w.Header().Set("Access-Control-Allow-Headers", "*")
// 	w.Header().Set("access-control-expose-headers", "*")
// 	w.Header().Set("Content-Type", "application/octet-stream")

// 	fn := strings.Split(header.Filename, ".")
// 	if len(fn) > 1 {
// 		fn[len(fn)-1] = "xlsx"
// 	}
// 	w.Header().Set("File-Name", strings.Join(fn, "."))
// 	w.Write(b)
// }

func apiHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("new request")
	file, header, err := r.FormFile("file")
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	b, err := ioutil.ReadAll(file)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	data := xmlUnMarshal(b)
	f := convertXlsx(data)

	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "*")
	w.Header().Set("access-control-expose-headers", "*")
	w.Header().Set("Content-Type", "application/octet-stream")

	fn := strings.Split(header.Filename, ".")
	if len(fn) > 1 {
		fn[len(fn)-1] = "xlsx"
	}
	w.Header().Set("File-Name", strings.Join(fn, "."))
	f.Write(w)
}

func xmlUnMarshal(b []byte) DataFormat {

	data := DataFormat{}
	err := xml.Unmarshal(b, &data)
	if nil != err {
		fmt.Println("Error unmarshalling from XML", err)
	}
	return data
}

func convertXlsx(d DataFormat) *xlsx.File {

	gtin := make([]string, 0)
	for _, product := range d.NFe.InfNFe.Det {
		if product.Prod.CEAN != "" {
			gtin = append(gtin, product.Prod.CEAN)
		}
	}
	f := xlsx.NewFile()

	var rst map[string]*sqlProduct
	if len(gtin) > 0 {
		connection, err := makeSQL(config.SQL.Host, config.SQL.Port, config.SQL.User, config.SQL.Password)
		if err != nil {
			log.Println(err)
			return f
		}

		where := " WHERE GTIN IN ('" + strings.Join(gtin, "', '") + "')"
		rst = connection.getCodBarras(where)
		fmt.Println(rst)
	}

	s, err := f.AddSheet("PLAN 1")
	if err != nil {
		log.Println(err)
		return f
	}
	r := s.AddRow()
	r.AddCell().SetString("Codigo_Produto_Nfe")
	r.AddCell().SetString("Codigo_Ean")
	r.AddCell().SetString("Desc_Produto")
	r.AddCell().SetString("NCM")
	r.AddCell().SetString("Quantidade")
	r.AddCell().SetString("Valor_Un")
	r.AddCell().SetString("Codigo_linx")
	r.AddCell().SetString("Desc_Produto_Linx")
	r.AddCell().SetString("Codigo_barras")

	for _, product := range d.NFe.InfNFe.Det {
		r = s.AddRow()
		r.AddCell().SetString(product.Prod.Cprod)
		r.AddCell().SetString(product.Prod.CEAN)
		r.AddCell().SetString(product.Prod.DescProduto)
		r.AddCell().SetString(product.Prod.NCM)
		r.AddCell().SetFloat(product.Prod.Quantidade)
		r.AddCell().SetFloat(product.Prod.ValorUni)
		if sqlitem, ok := rst[product.Prod.CEAN]; ok {
			r.AddCell().SetString(fmt.Sprintf("%s.%s.%02d", sqlitem.Produto, sqlitem.CorProduto, sqlitem.Tamanho))
			r.AddCell().SetString(fmt.Sprintf("%s %s %s", sqlitem.DescProduto, sqlitem.DescColorProd, sqlitem.Grade))
			r.AddCell().SetString(sqlitem.Codbarra)

		}
	}

	return f

}
