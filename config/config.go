package config

type PropriedadesDB struct {
	Port                  string `env:"MY_APP_PORT" env-default:"5001"`
	Host                  string `env:"HOST" env-default:"localhost"`
	DBHost                string `env:"DB_HOST" env-default:"localhost"`
	DBPort                string `env:"DB_PORT" env-default:"27017"`
	DBName                string `env:"DB_NAME" env-default:"desafio"`
	AlunosCollection      string `env:"COLLECTION_NAME" env-default:"alunos"`      //nome da coleção
	ProfessoresCollection string `env:"COLLECTION_NAME" env-default:"professores"` //nome da coleção
	CursosCollection      string `env:"COLLECTION_NAME" env-default:"cursos"`
	DisciplinasCollection string `env:"COLLECTION_NAME" env-default:"disciplinas"`
}
