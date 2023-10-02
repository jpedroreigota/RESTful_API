package handlers

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/url"

	"github.com/krunal4amity/tronicscorp/dbiface"
	"github.com/labstack/echo/v4"
	"github.com/labstack/gommon/log"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Alunos struct {
	ID primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"` /*onitempty serve para caso o
	campo não tenha sido preenchido, não haverá nenhum valor padrão, será ignorado*/
	Matricula int    `json:"matricula" bson:"matricula" validate:"required"`
	Nome      string `json:"nome" bson:"nome" validate:"required,max=20"`
	Sobrenome string `json:"sobrenome" bson:"sobrenome" validate:"required,max=20"`
	Telefone  int    `json:"telefone" bson:"telefone" validate:"required,max=10"`
}

type AlunosHandler struct {
	Col dbiface.Collection
}

func buscarAlunos(ctx context.Context, q url.Values, collection dbiface.Collection) ([]Alunos, *echo.HTTPError) {
	var alunos []Alunos
	filter := make(map[string]interface{})
	for k, v := range q {
		filter[k] = v[0]
	}
	if filter["_id"] != nil { //convertendo o id, que no filter é um string, para um primitiveObjectID
		docID, err := primitive.ObjectIDFromHex(filter["_id"].(string))
		if err != nil {
			return alunos, echo.NewHTTPError(http.StatusInternalServerError, "Unable to convert to ObjectID")
		}
		filter["_id"] = docID
	}
	cursor, err := collection.Find(ctx, bson.M(filter)) /*filter, que corresponde ao q url.Values, está sendo
	convertido para bson. Filter é argumento do método Find e está sendo definido logo acima*/
	if err != nil {
		log.Errorf("Unable to find the student: %v", err)
		return alunos, echo.NewHTTPError(http.StatusNotFound, "Unable to find the student")
	}
	err = cursor.All(ctx, &alunos)
	if err != nil {
		log.Errorf("Unable to read the cursor: %v", err)
		return alunos, echo.NewHTTPError(http.StatusBadRequest, "Unable to parse request payload")
	}
	return alunos, nil
}

func (h *AlunosHandler) BuscarAlunos(c echo.Context) error {
	alunos, err := buscarAlunos(context.Background(), c.QueryParams(), h.Col) /*c.QueryParams() para especificar
	consultas, por exemplo, caso não queira fazer um GET de todos os produtos, mas apenas de um produto com um
	determinado nome*/
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, alunos)
}

func buscarAluno(ctx context.Context, id string, collection dbiface.Collection) (Alunos, *echo.HTTPError) {
	var alunos Alunos
	docID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return alunos, echo.NewHTTPError(http.StatusInternalServerError, "Unable to convert to ObjectID")
	}
	filter := bson.M{"_id": docID}
	res := collection.FindOne(ctx, filter)
	err = res.Decode(&alunos)
	if err != nil {
		return alunos, echo.NewHTTPError(http.StatusNotFound, "Unable to find the student")
	}
	return alunos, nil
}

func (h *AlunosHandler) BuscarAluno(c echo.Context) error {
	alunos, err := buscarAluno(context.Background(), c.Param("id"), h.Col)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, alunos)
}

func inserirAluno(ctx context.Context, alunos []Alunos, collection dbiface.Collection) ([]interface{}, *echo.HTTPError) {
	var insertedIds []interface{}
	for _, aluno := range alunos {
		aluno.ID = primitive.NewObjectID()
		insertID, err := collection.InsertOne(ctx, aluno)
		if err != nil {
			log.Errorf("Unable to insert :%v", err)
			return nil, echo.NewHTTPError(http.StatusInternalServerError, "Unable to connect to database")
		}
		insertedIds = append(insertedIds, insertID.InsertedID)
	}
	return insertedIds, nil
}

func (h *AlunosHandler) InserirAluno(c echo.Context) error {
	var alunos []Alunos
	c.Echo().Validator = &AlunosValidator{validator: v}
	if err := c.Bind(&alunos); err != nil {
		log.Errorf("Unable to bind: %v", err)
		return c.JSON(http.StatusBadRequest, "Unable to parse request payload")
	}
	IDs, err := inserirAluno(context.Background(), alunos, h.Col)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusCreated, IDs)
}

func alterarAluno(ctx context.Context, id string, reqBody io.ReadCloser, collection dbiface.Collection) (Alunos, *echo.HTTPError) {
	var alunos Alunos

	//convertendo o id, que é um string, para primitive.ObjectID
	docID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		log.Errorf("Cannot convert to ObjectID: %v", err)
		return alunos, echo.NewHTTPError(http.StatusInternalServerError, "Unable to convert to ObjectID")
	}

	//procurar se o aluno existe, caso contrário erro 404 (Not Found)
	filter := bson.M{"_id": docID}
	res := collection.FindOne(ctx, filter)
	if err := res.Decode(&alunos); err != nil {
		log.Errorf("Unable to decode to student: %v", err)
		return alunos, echo.NewHTTPError(http.StatusNotFound, "Unable to find the student")
	} /*todos os valores, obtidos do método FindOne e atribuídos para res, são decodificados para alunos,
	ou seja, alunos está sendo populado*/

	//JSON decodificação do reqBody
	if err := json.NewDecoder(reqBody).Decode(&alunos); err != nil {
		log.Errorf("Unable to decode using reqBody: %v", err)
		return alunos, echo.NewHTTPError(http.StatusBadRequest, "Unable to parse request payload")
	} /*ler o reqBody e usar as informações inseridas para popular os campos de alunos, que "representa" a
	struct Alunos*/

	//atualização do aluno
	_, err = collection.UpdateOne(ctx, filter, bson.M{"$set": alunos}) /* _, err, pois UpdateOne possui 2
	retornos, *mongo.UpdateResult e error, sendo que não necessita-se aqui do UpdateResult*/
	if err != nil {
		log.Errorf("Unable to update the student: %v", err)
		return alunos, echo.NewHTTPError(http.StatusInternalServerError, "Unable to update the student")
	}
	return alunos, nil
}

func (h *AlunosHandler) AtualizarAluno(c echo.Context) error {
	alunos, err := alterarAluno(context.Background(), c.Param("id"), c.Request().Body, h.Col) /*criando o
	método alterarAluno*/
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, alunos)
}

func deletarAluno(ctx context.Context, id string, collection dbiface.Collection) (int64, *echo.HTTPError) {
	docID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		log.Errorf("Unable to delete the student: %v", err)
		return 0, echo.NewHTTPError(http.StatusInternalServerError, "Unable to convert to ObjectID")
	}
	res, err := collection.DeleteOne(ctx, bson.M{"_id": docID})
	if err != nil {
		return 0, echo.NewHTTPError(http.StatusNotFound, "Unable to delete the student")
	}
	return res.DeletedCount, nil
}

func (h *AlunosHandler) DeletarAluno(c echo.Context) error {
	delCount, err := deletarAluno(context.Background(), c.Param("id"), h.Col)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, delCount)
}
