package main

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
	"github.com/smahs/go-git-server/storage"

	"gopkg.in/src-d/go-git.v4/plumbing/protocol/packp"
	"gopkg.in/src-d/go-git.v4/plumbing/transport"
	plumber "gopkg.in/src-d/go-git.v4/plumbing/transport/server"
)

func servicePacket(service string) []byte {
	var term, flushPkt = "\n", "0000"
	var packet = fmt.Sprintf("# service=git-%s%s", service, term)
	s := strconv.FormatInt(int64(len(packet)+4), 16)
	if len(s)%4 != 0 {
		s = strings.Repeat("0", 4-len(s)%4) + s
	}
	return []byte(s + packet + flushPkt)
}

func initMux() http.Handler {
	var router = mux.NewRouter()
	router.Methods("GET").
		Path("/{owner}/{repo}/info/refs").
		Queries("service", "git-{service}").
		HandlerFunc(getAdvRefs)
	router.Methods("POST").
		Path("/{owner}/{repo}/git-{service}").
		HeadersRegexp("content-type",
			"application/x-git-(upload|receive)-pack-request").
		HandlerFunc(servicePacks)
	return router
}

func gitEndpoint(path string) (*transport.Endpoint, error) {
	var uri = &url.URL{
		Scheme: addr.Scheme,
		Host:   addr.Host,
		Path:   path,
	}
	return transport.NewEndpoint(uri.String())
}

func getAdvRefs(w http.ResponseWriter, r *http.Request) {
	var (
		vars     = mux.Vars(r)
		service  = vars["service"]
		repoName = fmt.Sprintf("/%s/%s", vars["owner"], vars["repo"])
		repo     = store.NewRepoStore(repoName)
		git      = plumber.NewServer(repo)
		refs     *packp.AdvRefs
		err      error
	)

	switch service {
	case "upload-pack":
		if refs, err = advRefsUpload(r, git); err != nil {
			goto handleError
		}
	case "receive-pack":
		if refs, err = advRefsReceive(r, git); err != nil {
			goto handleError
		}
	default:
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type",
		fmt.Sprintf("application/x-git-%s-advertisement", service))
	w.Header().Set("Cache-Control", "no-cache")
	w.Write(servicePacket(service))
	err = refs.Encode(w)

handleError:
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func advRefsReceive(r *http.Request,
	git transport.Transport) (*packp.AdvRefs, error) {
	var ep, _ = gitEndpoint(r.URL.Path)
	var sess, err = git.NewReceivePackSession(ep, nil)
	if err != nil {
		return nil, err
	}
	return sess.AdvertisedReferences()
}

func advRefsUpload(r *http.Request,
	git transport.Transport) (*packp.AdvRefs, error) {
	var ep, _ = gitEndpoint(r.URL.Path)
	var sess, err = git.NewUploadPackSession(ep, nil)
	if err != nil {
		return nil, err
	}
	return sess.AdvertisedReferences()
}

func servicePacks(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	var (
		vars     = mux.Vars(r)
		service  = vars["service"]
		repoName = fmt.Sprintf("/%s/%s", vars["owner"], vars["repo"])
		repo     = store.NewRepoStore(repoName)
	)

	switch service {
	case "receive-pack":
		receivePack(w, r, repo)
	case "upload-pack":
		uploadPack(w, r, repo)
	default:
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type",
		fmt.Sprintf("application/x-git-%s-result", service))
}

func receivePack(w http.ResponseWriter, r *http.Request,
	repo *storage.RepoStore) {
	var (
		ep, _  = gitEndpoint(r.URL.Path)
		git    = plumber.NewServer(repo)
		req    = packp.NewReferenceUpdateRequest()
		sess   transport.ReceivePackSession
		status *packp.ReportStatus
		err    error
	)

	if sess, err = git.NewReceivePackSession(ep, nil); err != nil {
		goto handleError
	}

	if err = req.Decode(r.Body); err != nil {
		goto handleError
	}

	if status, err = sess.ReceivePack(r.Context(), req); err != nil {
		goto handleError
	}

	if err = status.Encode(w); err != nil {
		goto handleError
	}

handleError:
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func uploadPack(w http.ResponseWriter, r *http.Request,
	repo *storage.RepoStore) {
	var (
		ep, _ = gitEndpoint(r.URL.Path)
		git   = plumber.NewServer(repo)
		req   = packp.NewUploadPackRequest()
		sess  transport.UploadPackSession
		resp  *packp.UploadPackResponse
		err   error
	)

	if sess, err = git.NewUploadPackSession(ep, nil); err != nil {
		goto handleError
	}

	if err = req.Decode(r.Body); err != nil {
		goto handleError
	}

	if resp, err = sess.UploadPack(r.Context(), req); err != nil {
		goto handleError
	}

	if err = resp.Encode(w); err != nil {
		goto handleError
	}

handleError:
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
	}
}
