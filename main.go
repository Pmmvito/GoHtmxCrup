package main

import (
	"html/template"
	"log"
	"net/http"
	"strconv"
	"strings"
)

func main() {
	appTemplates := NewAppTemplates()

	http.HandleFunc("/", appTemplates.homeHandler)
	http.HandleFunc("/bloco-register", appTemplates.registerHandler)
	http.HandleFunc("/bloco-login", appTemplates.loginHandler)

	http.HandleFunc("/salvar-form", appTemplates.salvarFormHandler)
	http.HandleFunc("/meus-forms", appTemplates.listarFormsHandler)
	http.HandleFunc("/editar-form", appTemplates.editarFormHandler)
	http.HandleFunc("/atualizar-form", appTemplates.atualizarFormHandler)
	http.HandleFunc("/deletar-form", appTemplates.deletarFormHandler)

	log.Println("Servidor iniciado em http://localhost:8080")
	http.ListenAndServe(":8080", nil)
}

type session struct {
	UserEmail string
}

type usuario struct {
	ID       int
	Name     string
	Email    string
	Password string
}

type formulario struct {
	ID        int
	UserEmail string
	Nome      string
	Texto     string
}

type AppTemplates struct {
	templates  *template.Template
	usuarios   map[string]*usuario
	sessions   map[string]*session
	forms      map[int]*formulario
	userID     int
	nextFormID int
}

func NewAppTemplates() *AppTemplates {
	temple, err := template.ParseGlob("Templates/*.html")
	if err != nil {
		log.Fatalf("Erro ao carregar templates: %v", err)
	}
	return &AppTemplates{
		templates:  temple,
		usuarios:   make(map[string]*usuario),
		sessions:   make(map[string]*session),
		forms:      make(map[int]*formulario),
		userID:     0,
		nextFormID: 0,
	}
}

func (t *AppTemplates) homeHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := t.templates.ExecuteTemplate(w, "layout.html", nil); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Printf("Erro ao executar template: %v", err)
	}
}

func (t *AppTemplates) loginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		email := strings.TrimSpace(strings.ToLower(r.FormValue("email")))
		password := r.FormValue("password")

		if user, exists := t.usuarios[email]; exists {
			if user.Password == password {
				t.sessions[email] = &session{UserEmail: email}

				data := map[string]interface{}{
					"Usuario": user,
				}
				w.Header().Set("Content-Type", "text/html; charset=utf-8")
				if err := t.templates.ExecuteTemplate(w, "bloco-formulario", data); err != nil {
					http.Error(w, "Erro ao renderizar formulário", http.StatusInternalServerError)
					log.Printf("Erro ao executar template: %v", err)
				}
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

func (t *AppTemplates) registerHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		name := r.FormValue("name")
		email := strings.ToLower(r.FormValue("email"))
		password := r.FormValue("password")

		if name == "" || email == "" || password == "" {
			w.Write([]byte(`<div class="alert alert-danger">Todos os campos são obrigatórios!</div>`))
			return
		}

		if _, existente := t.usuarios[email]; existente {
			w.Write([]byte(`<div class="alert alert-warning">Este email já está cadastrado!</div>`))
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
		w.Write([]byte(`<div class="alert alert-success">Usuário cadastrado com sucesso!</div>`))
		return
	}
}

func (t *AppTemplates) salvarFormHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		return
	}

	email := strings.TrimSpace(strings.ToLower(r.FormValue("email")))
	if _, ok := t.sessions[email]; !ok {
		http.Error(w, "Sessão inválida", http.StatusUnauthorized)
		return
	}

	nome := strings.TrimSpace(r.FormValue("nome"))
	texto := strings.TrimSpace(r.FormValue("texto"))

	t.nextFormID++
	t.forms[t.nextFormID] = &formulario{
		ID:        t.nextFormID,
		UserEmail: email,
		Nome:      nome,
		Texto:     texto,
	}

	t.renderListaForms(w, email)
}

func (t *AppTemplates) listarFormsHandler(w http.ResponseWriter, r *http.Request) {
	email := strings.TrimSpace(strings.ToLower(r.FormValue("email")))
	t.renderListaForms(w, email)
}

func (t *AppTemplates) renderListaForms(w http.ResponseWriter, email string) {
	var list []*formulario
	for _, f := range t.forms {
		if f.UserEmail == email {
			list = append(list, f)
		}
	}

	data := map[string]interface{}{"Forms": list}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := t.templates.ExecuteTemplate(w, "bloco-lista-forms", data); err != nil {
		http.Error(w, "Erro ao renderizar lista", http.StatusInternalServerError)
	}
}

func (t *AppTemplates) editarFormHandler(w http.ResponseWriter, r *http.Request) {
	email := strings.TrimSpace(strings.ToLower(r.FormValue("email")))
	if _, ok := t.sessions[email]; !ok {
		http.Error(w, "Sessão inválida", http.StatusUnauthorized)
		return
	}

	id, _ := strconv.Atoi(r.URL.Query().Get("id"))
	if f, ok := t.forms[id]; ok && f.UserEmail == email {
		data := map[string]interface{}{
			"Form":    f,
			"Usuario": t.usuarios[email],
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		t.templates.ExecuteTemplate(w, "bloco-form-editar", data)
	}
}

func (t *AppTemplates) atualizarFormHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		return
	}

	email := strings.TrimSpace(strings.ToLower(r.FormValue("email")))
	if _, ok := t.sessions[email]; !ok {
		return
	}

	id, _ := strconv.Atoi(r.FormValue("id"))
	if f, ok := t.forms[id]; ok && f.UserEmail == email {
		f.Nome = r.FormValue("nome")
		f.Texto = r.FormValue("texto")
	}

	data := map[string]interface{}{"Usuario": t.usuarios[email]}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	t.templates.ExecuteTemplate(w, "bloco-formulario", data)
}

func (t *AppTemplates) deletarFormHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		return
	}

	email := strings.TrimSpace(strings.ToLower(r.FormValue("email")))
	if _, ok := t.sessions[email]; !ok {
		return
	}

	id, _ := strconv.Atoi(r.FormValue("id"))
	if f, ok := t.forms[id]; ok && f.UserEmail == email {
		delete(t.forms, id)
	}

	t.renderListaForms(w, email)
}
