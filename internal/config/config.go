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
	Mode          string `json:"mode"`
	MongoDBConfig `json:"mongo_config"`
	RMQConfig     RMQConfig `json:"rmq_config"`
}

type MongoDBConfig struct {
	MDB_URI              string            `json:"mdb_uri"`
	MDB_NAME             string            `json:"mdb_name"`
	MDB_CLIENT           string            `json:"mdb_client"`
	MDB_DELIVERY_ADDRESS string            `json:"mdb_delivery_address"`
	MDB_GRIFE            string            `json:"mdb_grife"`
	MDB_ORDER            string            `json:"mdb_order"`
	MDB_DET_ORDER        string            `json:"mdb_det_order"`
	MDB_PAYMENT          string            `json:"mdb_payment"`
	MDB_PRODUTC          string            `json:"mdb_product"`
	MDB_SUPPLIER         string            `json:"mdb_supplier"`
	MDB_USER             string            `json:"mdb_user"`
	MDB_COLLECTIONS      map[string]string `json:"mdb_collections"`
}

type RMQConfig struct {
	RMQ_URI                  string `json:"rmq_uri"`
	RMQ_MAXX_RECONNECT_TIMES int    `json:"rmq_maxx_reconnect_times"`
}

func NewConfig() *Config {

	conf := defaultConf()

	SRV_MODE := os.Getenv("SRV_MODE")
	if SRV_MODE != "" {
		conf.Mode = SRV_MODE
	}

	SRV_MDB_URI := os.Getenv("SRV_MDB_URI")
	if SRV_MDB_URI != "" {
		conf.MDB_URI = SRV_MDB_URI
	}

	SRV_MDB_NAME := os.Getenv("SRV_MDB_NAME")
	if SRV_MDB_NAME != "" {
		conf.MDB_NAME = SRV_MDB_NAME
	}

	SRV_MDB_COLLECTIONS := os.Getenv("SRV_MDB_COLLECTIONS")
	if SRV_MDB_COLLECTIONS != "" {
		collectionsMap := parseCollectionsString(SRV_MDB_COLLECTIONS)
		conf.MDB_COLLECTIONS = collectionsMap
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
		MongoDBConfig: MongoDBConfig{
			MDB_URI:         "mongodb://admin:supersenha@localhost:27017/",
			MDB_NAME:        "teste_db",
			MDB_COLLECTIONS: make(map[string]string),
		},

		RMQConfig: RMQConfig{
			RMQ_URI: "amqp://admin:supersenha@localhost:5672/",
		},
	}
	// Adicione as coleções padrão ao mapa MDB_COLLECTIONS
	defaultCollections := "meiospagamentos, categorias, clientes, fornecedores, orcamentos, produtos, compras"
	collectionsMap := parseCollectionsString(defaultCollections)
	default_conf.MongoDBConfig.MDB_COLLECTIONS = collectionsMap

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
