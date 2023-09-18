package di_test

import (
	"testing"

	"github.com/livebud/bud/pkg/di"
	"github.com/matryer/is"
)

type Context struct {
	Env *Env
	Log *Log
}

func TestUnmarshal(t *testing.T) {
	is := is.New(t)
	in := di.New()
	di.Provide[*Env](in, loadEnv)
	di.Provide[*Log](in, loadLog)
	var ctx Context
	err := di.Unmarshal(in, &ctx)
	is.NoErr(err)
	is.Equal(ctx.Env.name, "production")
	is.Equal(ctx.Log.lvl, "info")
}

func TestUnmarshalPointer(t *testing.T) {
	is := is.New(t)
	in := di.New()
	di.Provide[*Env](in, loadEnv)
	di.Provide[*Log](in, loadLog)
	var ctx Context
	err := di.Unmarshal(in, &ctx)
	is.NoErr(err)
	is.Equal(ctx.Env.name, "production")
	is.Equal(ctx.Log.lvl, "info")
}

type Session[Data any] struct {
	Data    Data
	Flashes []string
}

func TestUnmarshalGeneric(t *testing.T) {
	type User struct {
		ID int
	}
	var Context struct {
		Session *Session[*User]
	}
	is := is.New(t)
	in := di.New()
	di.Provide[*Session[*User]](in, func() (s *Session[*User]) {
		return &Session[*User]{
			Data: &User{ID: 1},
		}
	})
	err := di.Unmarshal(in, &Context)
	is.NoErr(err)
	is.Equal(Context.Session.Data.ID, 1)
}