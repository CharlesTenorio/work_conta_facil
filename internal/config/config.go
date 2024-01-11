package config

import (
	"log"
	"os"
	"strings"
)

const (
	DEVELOPER    = "developer"
	HOMOLOGATION = "homologation"
	PRODUCTION   = "production"
)

type Config struct {
	Mode      string       `json:"mode"`
	MDBConfig MongoConfig  `json:"mdb_config"`
	RMQConfig RabbitConfig `json:"rmq_config"`
}

type MongoConfig struct {
	DB_URI          string            `json:"mdb_uri"`
	DB_NAME         string            `json:"mdb_name"`
	MDB_COLLECTIONS map[string]string `json:"mdb_collections"`
}

type RabbitConfig struct {
	RMQ_URI string `json:"rmq_uri"`
}

func NewConfig() *Config {

	conf := defaultConf()

	SRV_MODE := os.Getenv("SRV_MODE")
	if SRV_MODE != "" {
		conf.Mode = SRV_MODE
	}

	SRV_MDB_URI := os.Getenv("SRV_MDB_URI")
	if SRV_MDB_URI != "" {
		conf.MDBConfig.DB_URI = SRV_MDB_URI
	} else {
		log.Println("environment variable SRV_MDB_URI not found")
		os.Exit(1)
	}

	SRV_MDB_NAME := os.Getenv("SRV_MDB_NAME")
	if SRV_MDB_NAME != "" {
		conf.MDBConfig.DB_NAME = SRV_MDB_NAME
	} else {
		log.Println("environment variable SRV_MDB_NAME not found")
		os.Exit(1)
	}

	SRV_RMQ_URI := os.Getenv("SRV_RMQ_URI")
	if SRV_RMQ_URI != "" {
		conf.RMQConfig.RMQ_URI = SRV_RMQ_URI
	} else {
		log.Println("environment variable SRV_RMQ_URI not found")
		os.Exit(1)
	}

	return conf
}

func defaultConf() *Config {
	default_conf := Config{
		Mode: PRODUCTION,
		MDBConfig: MongoConfig{
			DB_URI:          "mongodb://admin:supersenha@localhost:27017/",
			DB_NAME:         "teste_db",
			MDB_COLLECTIONS: make(map[string]string),
		},
		RMQConfig: RabbitConfig{
			RMQ_URI: "amqp://admin:supersenha@localhost:5672/",
		},
	}
	// Adicione as coleções padrão ao mapa MDB_COLLECTIONS
	defaultCollections := "meiospagamentos, categorias, clientes, fornecedores, orcamentos, produtos, compras"
	collectionsMap := parseCollectionsString(defaultCollections)
	default_conf.MDBConfig.MDB_COLLECTIONS = collectionsMap

	return &default_conf
}

func parseCollectionsString(collectionsString string) map[string]string {
	collections := make(map[string]string)

	// Separar a string usando a vírgula como delimitador
	collectionNames := strings.Split(collectionsString, ",")

	// Adicionar cada nome de coleção ao mapa
	for _, name := range collectionNames {
		// Remover espaços em branco ao redor do nome da coleção
		name = strings.TrimSpace(name)
		collections[name] = name
	}

	return collections
}
