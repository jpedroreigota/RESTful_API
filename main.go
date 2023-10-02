package main

import (
	"context"
	"fmt"

	"github.com/ilyakaznacheev/cleanenv"
	"github.com/krunal4amity/tronicscorp/config"
	"github.com/krunal4amity/tronicscorp/handlers"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/labstack/gommon/log"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	c              *mongo.Client
	db             *mongo.Database
	col            *mongo.Collection
	alunosCol      *mongo.Collection
	professoresCol *mongo.Collection
	cursosCol      *mongo.Collection
	disciplinasCol *mongo.Collection
	cfg            config.PropriedadesDB
)

func init() {
	err := cleanenv.ReadEnv(&cfg)
	if err != nil {
		log.Fatalf("Configuration cannot be read: %v", err)
	}
	connectURI := fmt.Sprintf("mongodb://%s:%s", cfg.DBHost, cfg.DBPort)
	c, err := mongo.Connect(context.Background(), options.Client().ApplyURI(connectURI))
	if err != nil {
		log.Fatalf("Unable to connect database: %v", err) /*em connectURI, por exemplo, caso o DBPort esteja
		errado, resultará neste erro*/
	}

	db = c.Database(cfg.DBName)
	alunosCol = db.Collection(cfg.AlunosCollection)
	professoresCol = db.Collection(cfg.ProfessoresCollection)
	cursosCol = db.Collection(cfg.CursosCollection)
	disciplinasCol = db.Collection(cfg.DisciplinasCollection)
} //responsável pela conexão com a API

func mensagemServidor(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		fmt.Println("estamos dentro")

		return next(c)
	}
}

func main() {
	e := echo.New()
	e.Logger.SetLevel(log.DEBUG)
	e.Use(middleware.Logger())  // Logger
	e.Use(middleware.Recover()) // Recover
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{echo.GET, echo.HEAD, echo.PUT, echo.PATCH, echo.POST, echo.DELETE},
	}))
	e.Pre(middleware.RemoveTrailingSlash())
	e.Pre(mensagemServidor)
	/*e.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
		Format: `$(time_rfc3339_nano) $(remote_ip) $(host) $(method) $(uri) $(user_agent) ` +
			`$(status) $(error) $(latency_human)` + "\n",
	}))*/

	h := &handlers.AlunosHandler{Col: alunosCol}
	uh := &handlers.ProfessoresHandler{Col: professoresCol}
	ah := &handlers.CursosHandler{Col: cursosCol}
	oh := &handlers.DisciplinasHandler{Col: disciplinasCol}

	e.POST("/alunos", h.InserirAluno, middleware.BodyLimit("1M"))
	e.GET("/alunos", h.BuscarAlunos)
	e.GET("/alunos/:id", h.BuscarAluno)
	e.PUT("/alunos/:id", h.AtualizarAluno, middleware.BodyLimit("1M"))
	e.DELETE("/alunos/:id", h.DeletarAluno)

	e.POST("/professores", uh.InserirProfessor, middleware.BodyLimit("1M"))
	e.GET("/professores", uh.BuscarProfessores)
	e.GET("/professores/:id", uh.BuscarProfessor)
	e.PUT("/professores/:id", uh.AtualizarProfessor, middleware.BodyLimit("1M"))
	e.DELETE("/professores/:id", uh.DeletarProfessor)

	e.POST("/cursos", ah.InserirCurso, middleware.BodyLimit("1M"))
	e.GET("/cursos", ah.BuscarCursos)
	e.GET("/cursos/:id", ah.BuscarCurso)
	e.PUT("/cursos/:id", ah.AtualizarCurso, middleware.BodyLimit("1M"))
	e.DELETE("/cursos/:id", ah.DeletarCurso)

	e.POST("/disciplinas", oh.InserirDisciplina, middleware.BodyLimit("1M"))
	e.GET("/disciplinas", oh.BuscarDisciplinas)
	e.GET("/disciplinas/:id", oh.BuscarDisciplina)
	e.PUT("/disciplinas/:id", oh.AtualizarDisciplina, middleware.BodyLimit("1M"))
	e.DELETE("/disciplinas/:id", oh.DeletarDisciplina)

	e.Logger.Print(fmt.Sprintf("Listening on port: %s", cfg.Port))
	e.Logger.Fatal(e.Start(fmt.Sprintf("%s:%s", cfg.Host, cfg.Port)))
}
