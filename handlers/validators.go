package handlers

import "gopkg.in/go-playground/validator.v9"

var (
	v = validator.New()
)

type AlunosValidator struct{
	validator *validator.Validate
}

func (a *AlunosValidator) Validate(i interface{}) error{
	return a.validator.Struct(i)
}

type ProfessoresValidator struct{
	validator *validator.Validate
}

func (p *ProfessoresValidator) Validate(i interface{}) error{
	return p.validator.Struct(i)
}

type CursosValidator struct{
	validator *validator.Validate
}

func (c *CursosValidator) Validate(i interface{}) error{
	return c.validator.Struct(i)
}

type DisciplinasValidator struct{
	validator *validator.Validate
}

func (d *DisciplinasValidator) Validate(i interface{}) error{
	return d.validator.Struct(i)
}
