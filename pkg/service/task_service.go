package service

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"time"

	"github.com/katana/worker/orcafacil-go/internal/config/logger"
	"github.com/katana/worker/orcafacil-go/internal/dto"
	"github.com/katana/worker/orcafacil-go/pkg/adapter/mongodb"
	"github.com/katana/worker/orcafacil-go/pkg/adapter/rabbitmq"
	"github.com/katana/worker/orcafacil-go/pkg/model"
	amqp "github.com/rabbitmq/amqp091-go"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const DATE_FORMAT = "2006-01-02 15:04:05"

type taskService struct {
	rbmq_conn rabbitmq.RabbitInterface
	db_conn   mongodb.MongoDBInterface
}

func NewTaskService(rbmq_conn rabbitmq.RabbitInterface, db_conn mongodb.MongoDBInterface) *taskService {
	return &taskService{
		rbmq_conn: rbmq_conn,
		db_conn:   db_conn,
	}
}

func (tmcs *taskService) LerPrdCotacao(callback func(msg *amqp.Delivery)) error {
	err := tmcs.rbmq_conn.Connect()
	if err != nil {
		logger.Error("deu ruim na conexao como RabbitMQ", err)
		return err
	}

	// Substitua "QUEUE_PRDS_PARA_COTACAO" pelo nome da fila desejada
	tmcs.rbmq_conn.Start("QUEUE_PRDS_PARA_COTACAO", callback)
	return nil
}

func (tmcs *taskService) ReceiveMessage(ctx context.Context, message *model.MessageConversator) error {
	collection := tmcs.db_conn.GetCollection("sgStore")

	filter := bson.D{
		{Key: "data_type", Value: "message_count"},
		{Key: "id_message_conversator", Value: message.IdMessageConversator},
	}

	values := bson.D{
		{Key: "id_message_conversator", Value: message.IdMessageConversator},
		{Key: "bot_id", Value: message.BotId},
		{Key: "user_id", Value: message.UserId},
		{Key: "platform", Value: message.Platform},
		{Key: "teams_flow", Value: message.TeamsFlow},
		{Key: "watson_flow", Value: message.WatsonFlow},
		{Key: "created_at", Value: message.CreatedAt},
	}

	opts := options.Update().SetUpsert(true)

	update := bson.D{{Key: "$set", Value: values}}

	_, err := collection.UpdateOne(ctx, filter, update, opts)
	if err != nil {
		log.Println("Error while updating data")
		return err
	}

	return nil
}

func (tmcs *taskService) consumerCallback(msg *amqp.Delivery) {

	log.Println("New MSG received")

	mc := &dto.OrcamentoFilaPrdFornecedor{}
	err := json.Unmarshal(msg.Body, mc)
	if err != nil {
		log.Println("Failed to read a consumer msg")
		log.Println(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = tmcs.ReceiveMessage(ctx, mc)
	if err != nil {
		log.Println("Failed to save msg")
		log.Println(err)

	}

	if err := msg.Ack(false); err != nil {
		log.Println("Erro to ACK MSG Status")
	} else {
		log.Println("MSG Status update success")
	}
}

func (tmcs *taskService) Run() {

	err := tmcs.rbmq_conn.Connect()
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}

	tmcs.rbmq_conn.Start("FINANCIAL_MESSAGE_COUNTER", tmcs.consumerCallback)
}
