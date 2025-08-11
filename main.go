package main

import (
	"html/template"
	"log"
	"net/http"
	"strings"
)



func main() {
	appTemplates := NewAppTemplates()

	http.HandleFunc("/", appTemplates.homeHandler)
	http.HandleFunc("/bloco-register", appTemplates.registerHandler)
	http.HandleFunc("/bloco-login", appTemplates.loginHandler)

	log.Println("Servidor iniciado em http://localhost:8080")
	http.ListenAndServe(":8080", nil)
	
}


type usuario struct {
    ID       int    
    Name     string 
    Email    string 
    Password string 
}

type AppTemplates struct {
    templates *template.Template
    usuarios  map[string]*usuario 
    userID    int                 
}

func NewAppTemplates() *AppTemplates {
    temple, err := template.ParseGlob("Templates/*.html")
    if err != nil {
        log.Fatalf("Erro ao carregar templates: %v", err)
    }

    return &AppTemplates{
        templates: temple,
        usuarios:  make(map[string]*usuario),
        userID:    0,
    }
}

func (t *AppTemplates) homeHandler(w http.ResponseWriter, r *http.Request) {
    err := t.templates.ExecuteTemplate(w, "index.html", nil)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        log.Printf("Erro ao executar template: %v", err)
    }
}
//bloco 1//
func (t *AppTemplates) loginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
        email := strings.TrimSpace(strings.ToLower(r.FormValue("email")))
        password := r.FormValue("password")

        if user, exists := t.usuarios[email]; exists {
            if user.Password == password {
                w.Write([]byte(`<div class="alert alert-success">Login realizado com sucesso! Bem-vindo, ` + user.Name + `!</div>`))
                return
            } else {
                w.Write([]byte(`<div class="alert alert-danger">Email ou senha incorretos!</div>`))
                return
            }
    } else {
            w.Write([]byte(`<div class="alert alert-danger">Usuário não encontrado! Faça seu cadastro primeiro.</div>`))
            return
        }
    }

}

//bloco 2//

func (t *AppTemplates) registerHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
        name :=r.FormValue("name")
        email := strings.ToLower(r.FormValue("email"))
        password := r.FormValue("password")

        if name == "" || email == "" || password == "" {
            w.Write([]byte(`<div class="alert alert-danger">Todos os campos são obrigatórios!</div>`))
            return
        }

        if _, existente := t.usuarios[email]; existente {
            w.Write([]byte(`<div class="alert alert-warning">Este email já está cadastrado! <button class="btn btn-sm btn-link" onclick="document.getElementById('login-tab').click();">Fazer login</button></div>`))
            return
        }

        t.userID++
        novoUsuario := &usuario{
            ID:       t.userID,
            Name:     name,
            Email:    email,
            Password: password,
        }

        t.usuarios[email] = novoUsuario

        log.Printf("Novo usuário registrado: ID=%d, Nome=%s, Email=%s", novoUsuario.ID, novoUsuario.Name, novoUsuario.Email)

        w.Write([]byte(`<div class="alert alert-success">Usuário cadastrado com sucesso! <button class="btn btn-sm btn-link" onclick="document.getElementById('login-tab').click();">Fazer login</button></div>`))
        return
    }

}