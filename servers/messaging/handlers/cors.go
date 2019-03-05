package handlers

import "net/http"

/* TODO: implement a CORS middleware handler, as described
in https://drstearns.github.io/tutorials/cors/ that responds
with the following headers to all requests:

  Access-Control-Allow-Origin: *
  Access-Control-Allow-Methods: GET, PUT, POST, PATCH, DELETE
  Access-Control-Allow-Headers: Content-Type, Authorization
  Access-Control-Expose-Headers: Authorization
  Access-Control-Max-Age: 600
*/

//Cors ? Who knows?
type Cors struct {
	handler http.Handler
}

//NewCors makes a new Cors
func NewCors(givenHandler http.Handler) *Cors {
	if givenHandler != nil {
		return &Cors{givenHandler}
	}
	return nil
}

func (c *Cors) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	if r.Method == http.MethodOptions {
		w.Header().Add("Access-Control-Allow-Origin", "*")
		w.Header().Add("Access-Control-Allow-Methods", "GET, PUT, POST, PATCH, DELETE")
		w.Header().Add("Access-Control-Allow-Headers", "Content-Type, Authorization")
		w.Header().Add("Access-Control-Expose-Headers", "Authorization")
		w.Header().Add("Access-Control-Max-Age", "600")
	}

	c.handler.ServeHTTP(w, r)

}
