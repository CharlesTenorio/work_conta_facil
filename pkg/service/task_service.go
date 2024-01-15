package service

import (
	"context"
	"encoding/json"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"os"
	"time"

	"github.com/katana/worker/orcafacil-go/internal/config/logger"
	"github.com/katana/worker/orcafacil-go/internal/dto"
	"github.com/katana/worker/orcafacil-go/pkg/adapter/mongodb"
	"github.com/katana/worker/orcafacil-go/pkg/adapter/rabbitmq"
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

func (tmcs *taskService) RelacionarPrdConFornecedor(ctx context.Context, prsOrcamento dto.ProdutoEnviadoParaFilaDeOrcamentoDTO) error {
	collection := tmcs.db_conn.GetCollection("fornecedores")

	// Crie um slice de IDs de produtos a serem usados no operador $in
	var produtos []string
	for _, produto := range prsOrcamento.Produtos {
		produtos = append(produtos, produto.Nome)
	}

	// Construa a consulta com o operador $in
	filter := bson.M{"produtos.nome": bson.M{"$in": produtos}}
	projection := bson.M{"_id": 1, "produtos": 1}

	// Execute a consulta
	cursor, err := collection.Find(ctx, filter, options.Find().SetProjection(projection))
	if err != nil {
		return err
	}
	defer cursor.Close(ctx)

	// Processar os resultados, se necessário
	for cursor.Next(ctx) {
		var fornec dto.FornecedorDto // Substitua "model.Fornecedor" pelo tipo real do documento
		if err = cursor.Decode(&fornec); err != nil {
			return err
		}

		err = tmcs.UpdateForncedorPrd(ctx, prsOrcamento.IdOrcamento.String(), &fornec)
		if err != nil {
			logger.Error("Erro ao chamar UpdateForncedorPrd", err)
			// Trate o erro conforme necessário
		}

		fornecedorJson := fornec.FornecedorDtoConvet()
		logger.Info(fornecedorJson)
		// Faça algo com o fornecedor encontrado

	}

	if err := cursor.Err(); err != nil {
		return err
	}

	return nil
}

func (tmcs *taskService) consumerCallback(msg *amqp.Delivery) {

	prdContacao := &dto.ProdutoEnviadoParaFilaDeOrcamentoDTO{}
	err := json.Unmarshal(msg.Body, prdContacao)
	if err != nil {
		logger.Error("Erro to Unmarshal MSG Status", err)
	}

	err = tmcs.RelacionarPrdConFornecedor(context.Background(), *prdContacao)
	if err != nil {
		logger.Error("Erro to RelacionarPrdConFornecedor", err)
	}

	if err := msg.Ack(false); err != nil {
		logger.Error("Erro to ACK MSG Status", err)
	} else {
		logger.Info("MSG Status update success")
	}
}

func (tmcs *taskService) Run() {

	err := tmcs.rbmq_conn.Connect()
	if err != nil {
		logger.Error("Erro to Connect in RabbitMQ Channel", err)
		os.Exit(1)
	}

	tmcs.rbmq_conn.Start("QUEUE_PRDS_PARA_COTACAO", tmcs.consumerCallback)
}

func (tmcs *taskService) UpdateForncedorPrd(ctx context.Context, ID string, fornecedorUpdate *dto.FornecedorDto) error {
	collection := tmcs.db_conn.GetCollection("orcamentos")

	opts := options.Update().SetUpsert(true)
	objectID, err := primitive.ObjectIDFromHex(ID)
	if err != nil {
		logger.Error("Error to parse ObjectIDFromHex", err)

	}
	filter := bson.D{
		{Key: "_id", Value: objectID},
	}

	// Criando um mapa para armazenar as atualizações dinâmicas
	updateFields := make(map[string]interface{})

	// Adicionando todos os campos que você deseja atualizar
	updateFields["updated_at"] = time.Now().Format(time.RFC3339)
	updateFields["status"] = "enviado para fornecedor(s)"

	// Adiciona o campo Fornecedores ao mapa de atualização
	updateFields["fornecedor.id"] = fornecedorUpdate.FornecedorID
	updateFields["fornecedor.produtos"] = fornecedorUpdate.Produtos

	// Criando a atualização dinâmica com os campos fornecidos
	var updateFieldsDoc bson.D
	for key, value := range updateFields {
		updateFieldsDoc = append(updateFieldsDoc, bson.E{Key: key, Value: value})
	}

	// Criando a atualização final
	update := bson.D{{Key: "$set", Value: updateFieldsDoc}}

	// Executando a atualização
	_, err = collection.UpdateOne(ctx, filter, update, opts)
	if err != nil {
		logger.Error("Error while updating data", err)
		return err
	}

	return nil
}
