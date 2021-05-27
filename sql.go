package main

import (
	"context"
	"database/sql"
	"fmt"
	"net/url"
	"strings"

	_ "github.com/denisenkom/go-mssqldb" //bblablalba
)

type sqlStr struct {
	url *url.URL
	db  *sql.DB
}
type sqlProduct struct {
	Produto       string `json:"PRODUT0,omitempty"`
	CorProduto    string `json:"COR_PRODUTO,omitempty"`
	Tamanho       int    `json:"TAMANHO,omitempty"`
	DescProduto   string `json:"DESC_PRODUTO,omitempty"`
	Grade         string `json:"GRADE,omitempty"`
	Codbarra      string `json:"CODIGO_BARRA,omitempty"`
	DescColorProd string `json:"DESC_COR_PRODUTO,omitempty"`
}

func (s *sqlStr) getCodBarras(where string) map[string]*sqlProduct {

	rst := make(map[string]*sqlProduct)
	sel := fmt.Sprintf("SELECT GTIN,PRODUTO,COR_PRODUTO,TAMANHO,DESC_PRODUTO,DESC_COR_PRODUTO,GRADE,CODIGO_BARRA FROM LINX_TBFG..RCS_APP_PRODUTO_GS1%s", where)
	fmt.Println(sel)
	rows, err := s.db.QueryContext(context.Background(), sel, nil)
	if err != nil {
		fmt.Println(err)
		return nil
	}

	var gtin string
	var gtinI, prod, colorprod, size, descprod, grade, codbarra, desccolorprod interface{}

	for rows.Next() {

		sqlProduct := sqlProduct{}

		if err := rows.Scan(&gtinI, &prod, &colorprod, &size, &descprod, &desccolorprod, &grade, &codbarra); err != nil {
			fmt.Println(err)
			continue
		}

		gtin, _ = gtinI.(string)
		gtin = strings.Trim(gtin, " ")

		if prodS, ok := prod.(string); ok {
			sqlProduct.Produto = strings.Trim(prodS, " ")
		}

		if colorprodS, ok := colorprod.(string); ok {
			sqlProduct.CorProduto = strings.Trim(colorprodS, " ")
		}

		if sizeI, ok := size.(int); ok {
			sqlProduct.Tamanho = sizeI
		}

		if descprodS, ok := descprod.(string); ok {
			sqlProduct.DescProduto = strings.Trim(descprodS, " ")
		}

		if gradeS, ok := grade.(string); ok {
			sqlProduct.Grade = strings.Trim(gradeS, " ")
		}

		if codbarraS, ok := codbarra.(string); ok {
			sqlProduct.Codbarra = strings.Trim(codbarraS, " ")
		}
		if desccolorprodS, ok := desccolorprod.(string); ok {
			sqlProduct.DescColorProd = strings.Trim(desccolorprodS, " ")
		}

		rst[gtin] = &sqlProduct
	}

	return rst
}

func makeSQL(host, port, username, password string) (*sqlStr, error) {
	s := &sqlStr{}
	s.url = &url.URL{
		Scheme:   "sqlserver",
		User:     url.UserPassword(username, password),
		Host:     fmt.Sprintf("%s:%s", host, port),
		RawQuery: url.Values{}.Encode(),
	}
	return s, s.connect()
}

func (s *sqlStr) connect() error {
	var err error
	if s.db, err = sql.Open("sqlserver", s.url.String()); err != nil {
		return err
	}
	return s.db.PingContext(context.Background())
}

func (s *sqlStr) disconnect() error {
	return s.db.Close()
}
