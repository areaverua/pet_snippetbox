package main

import (
	"database/sql"
	"flag"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"html/template"
	"github.com/areaverua/snippetbox/pkg/models/mysql"
	_ "github.com/go-sql-driver/mysql"
)


type application struct{
	errorLog *log.Logger
	infoLog  *log.Logger
	snippets *mysql.SnippetModel
	templateCache map[string]*template.Template
}




func main(){

	addr := flag.String("addr", ":4000", "Сетевой адрес HTTP")
	dsn := flag.String("dsn", "web:pass@/snippetbox?parseTime=true", "Название MySQL источника данных")
	flag.Parse()


	infolog := log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)
	errorLog := log.New(os.Stderr, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)


	db, err := openDB(*dsn)
	if err != nil{
		errorLog.Fatal(err)
	}
	defer db.Close()

	templateCache, err := newTemplateCache("./ui/html/")
    if err != nil {
        errorLog.Fatal(err)
    }


	app := &application{
		errorLog: errorLog,
		infoLog: infolog,
		snippets: &mysql.SnippetModel{DB: db},
		templateCache: templateCache,
	}


	/*fileServer := http.FileServer(neuteredFileSystem{http.Dir("./static")})
	mux.Handle("/static", http.NotFoundHandler())
	mux.Handle("/static/", http.StripPrefix("/static", fileServer)) */


	srv := &http.Server{
		Addr: *addr,
		ErrorLog: errorLog,
		Handler: app.routes(),
	}


	infolog.Printf("Запуск сервера на %s", *addr)
	err = srv.ListenAndServe()
	errorLog.Fatal(err)
}


func openDB(dsn string) (*sql.DB, error){
	db, err := sql.Open("mysql", dsn)
	if err != nil{
		return nil, err
	}
	if err = db.Ping(); err != nil{
		return nil, err
	}

	return db, nil
}




type neuteredFileSystem struct{
	fs http.FileSystem
}


func (nfs neuteredFileSystem) Open(path string) (http.File, error){
	f, err := nfs.fs.Open(path)
	if err != nil {
		return nil, err
	}


	s, err := f.Stat()
	if s.IsDir(){
		index := filepath.Join(path, "index.html")
		if _, err := nfs.fs.Open(index); err != nil {
			closeErr := f.Close()
			if closeErr != nil{
				return nil, closeErr
			}
			return nil, err
		}
	}

	return f, nil
}