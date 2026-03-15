//routes.go

package Routes

import (
  "github.com/kataras/iris"
  "github.com/kataras/iris/hero"
  "os"

  "server/Controllers"
  "server/Datasource"
  "server/Services"
  "server/repository"
)

func Configure(app *iris.Application) {
  if os.Getenv("SKIP_MYSQL") == "1" {
    hero.Register(Services.NewRAGMemoriseService())
  } else {
    db := Datasource.GetInstace().GetMysqlDB()
    hero.Register(
      Services.NewMemoriseService(
        repository.NewMemoriseRepo(db),
      ),
    )
  }

  app.Get("/", Controllers.GetIndexHandler)

  app.PartyFunc("/", func(r iris.Party) {
    r.Post("Add", hero.Handler(Controllers.Add))
    r.Post("Reply", hero.Handler(Controllers.Reply))
    r.Post("Forget", hero.Handler(Controllers.Forget))
    r.Post("Status", hero.Handler(Controllers.Status))
  })

  app.PartyFunc("/marisa", func(r iris.Party) {
    r.Post("Add", hero.Handler(Controllers.Add))
    r.Post("Reply", hero.Handler(Controllers.Reply))
    r.Post("Forget", hero.Handler(Controllers.Forget))
    r.Post("Status", hero.Handler(Controllers.Status))
  })

app.PartyFunc("/api", func(r iris.Party) {
    r.Get("/stats", hero.Handler(Controllers.GetStats))
    r.Post("/auth/register", hero.Handler(Controllers.Register))
    r.Post("/auth/login", hero.Handler(Controllers.Login))
    r.Get("/messages", hero.Handler(Controllers.ListMessages))
    r.Post("/messages", hero.Handler(Controllers.PostMessage))
    r.Get("/affinity/me", hero.Handler(Controllers.AffinityMe))
r.Get("/affinity/logs", hero.Handler(Controllers.AffinityLogs))

})

app.PartyFunc("/api/custom", func(r iris.Party) {
  r.Get("/personas", hero.Handler(Controllers.ListPersonas))
  r.Post("/personas", hero.Handler(Controllers.CreatePersona))
  r.Delete("/personas/{id:uint64}", hero.Handler(Controllers.DeletePersona))

  r.Get("/personas/{id:uint64}/teach", hero.Handler(Controllers.ListPersonaTeach))
  r.Post("/personas/{id:uint64}/teach", hero.Handler(Controllers.AddPersonaTeach))
  r.Delete("/personas/{id:uint64}/teach/last", hero.Handler(Controllers.DeleteTeachLast))

  r.Post("/reply", hero.Handler(Controllers.CustomReply))

  r.Get("/personas/{id:uint64}/logs", hero.Handler(Controllers.ListPersonaLogs))
r.Delete("/personas/{id:uint64}/logs", hero.Handler(Controllers.ClearPersonaLogs))

r.Get("/logs", hero.Handler(Controllers.ListUserLogs))
// 可选
// r.Delete("/logs", hero.Handler(Controllers.ClearUserLogs))

})

app.PartyFunc("/api/classic", func(r iris.Party) {
  r.Get("/status", hero.Handler(Controllers.ClassicStatus))
  r.Get("/list", hero.Handler(Controllers.ClassicList))   // ✅ 新增
  r.Post("/reply", hero.Handler(Controllers.ClassicReply))
  r.Post("/teach", hero.Handler(Controllers.ClassicTeach))
  r.Post("/forget", hero.Handler(Controllers.ClassicForget))
})

}
