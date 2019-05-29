package main

import (
	"context"
	"io"
	"log"
)

// RegistratorWithLog implements Registrator that is instrumented with logging
type RegistratorWithLog struct {
	stdlog, errlog *log.Logger
	base           Registrator
}

// NewRegistratorWithLog instruments an implementation of the Registrator with simple logging
func NewRegistratorWithLog(base Registrator, stdout, stderr io.Writer) RegistratorWithLog {
	return RegistratorWithLog{
		base:   base,
		stdlog: log.New(stdout, "", log.LstdFlags),
		errlog: log.New(stderr, "", log.LstdFlags),
	}
}

// Register implements Registrator
func (rl RegistratorWithLog) Register(ctx context.Context, f *Form) (up1 *User, err error) {
	params := []interface{}{"RegistratorWithLog: calling Register with params:", ctx, f}
	rl.stdlog.Println(params...)
	defer func() {
		results := []interface{}{"RegistratorWithLog: Register return results:", up1, err}
		if err != nil {
			rl.errlog.Println(results...)
		} else {
			rl.stdlog.Println(results...)
		}
	}()
	return rl.base.Register(ctx, f)
}
