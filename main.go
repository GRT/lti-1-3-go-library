package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"github.com/GRT/lti-1-3-go-library/lti"
	"github.com/GRT/lti-1-3-go-library/ltiCache"
	"github.com/GRT/lti-1-3-go-library/registrationDatastore"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/gorilla/securecookie"
	"github.com/gorilla/sessions"
)

const (
	defaultPort               = 8345
	debugFlag                 = true
	sessionCookieName         = "lti1_3_zop"
	exampleLoginURL           = "/example/login"
	exampleLaunchURL          = "/example/launch"
	exampleMembersURL         = "/example/members"
	exampleGradeURL           = "/example/grade"
	exampleGradesURL          = "/example/grades"
	examplePayloadTemplateStr = `
		<html><head>
		<script>
		var myMembers = [];
		var memberMap = {};
		var myGrades = [];

		function fetchMembers() {
			var gradeamember = '<table><tr><th>Name</th><th>Roles</th><th>Action</th></tr>';
			var nuURL = window.location.href.split('/').slice(0,-1).join('/') + '/{{.MemberPathPart}}?launchId={{.LaunchID}}';
			document.getElementById('memberelement').innerHTML = 'fetching members';
			console.log('location:', window.location.href);
			console.log('nuURL:', nuURL);
			var request = new XMLHttpRequest();
			request.open('GET', nuURL, true);
			request.onload = function() {
				console.log('response status:', request.status, 'response:',this.response)
				var obj = JSON.parse(this.response)
				var payload = JSON.stringify(obj, undefined, 2);
				console.log('payload:', payload);
				document.getElementById('memberelement').innerHTML = payload;
				// copy the members (screen out blanks)
				for (i = 0; i < obj.members.length; ++i) {
					if(obj.members[i].user_id && obj.members[i].user_id !== "") {
						var idx = myMembers.length
						myMembers.push(obj.members[i])
						memberMap[obj.members[i].user_id] = obj.members[i];
						gradeamember += "\n<tr><td>" + (idx+1) + ") " + myMembers[idx].name + "</td><td>" + myMembers[idx].roles.join(",") + "</td><td>Grade: <input type=\"text\" size=\"3\" id=\"grade" + idx + "\" /> <button onclick=\"doGrade(" + idx + ")\">Send Grade</button> </td></tr>";
					}
				}
				console.log("member map: " + JSON.stringify(memberMap));
				gradeamember += "\n</table>";
				document.getElementById('gradeamember').innerHTML = gradeamember;
			}
			request.send();
		}
		function doGrade(idx) {
			var grade = document.getElementById("grade" + idx).value;
			console.log('grade index:', idx, 'with grade:', grade);
			if(!isValidGrade(grade)) {
				alert('Grade: "' + grade + '" is not a valid grade')
				return
			}
			if(idx>=0 && idx<myMembers.length) {
				// send the grade
				sendGrade(myMembers[idx], grade)
				document.getElementById("grade" + idx).value = ''
			} else {
				console.log('no grade for unknown member index:', idx)
			}
		}

		function sendGrade(member, score)  {
			var nuURL = window.location.href.split('/').slice(0,-1).join('/') + '/{{.ScorePathPart}}?launchId={{.LaunchID}}&score=' + score + '&userId=' + member.user_id;
			document.getElementById('scoreelement').innerHTML = 'submit score';
			console.log('location:', window.location.href);
			console.log('nuURL:', nuURL);
			var request = new XMLHttpRequest();
			request.open('GET', nuURL, true);
			request.onload = function() {
				console.log('sendGrade response status:', request.status, 'response:',this.response)
				var obj = JSON.parse(this.response)
				var payload = JSON.stringify(obj, undefined, 2);
				console.log('sendGrade payload:', payload);
				document.getElementById('scoreelement').innerHTML = payload;
				alert("Grade Response: " + this.response)
			}
			request.send();
		}
		function isValidGrade(str) {
			return /^\+?(0|[1-9]\d*)$/.test(str);
		}
		

		function fetchGrades() {
			myGrades = [];
			var nuURL = window.location.href.split('/').slice(0,-1).join('/') + '/{{.GradesPathPart}}?launchId={{.LaunchID}}';
			var gradeselement = document.getElementById('gradeselement');
			gradeselement.innerHTML = 'fetching grades';
			var gradelist = document.getElementById('gradelist');
			gradelist.innerHTML = '';
			console.log('location:', window.location.href);
			console.log('nuURL:', nuURL);
			var request = new XMLHttpRequest();
			request.open('GET', nuURL, true);
			request.onload = function() {
			console.log('response status for fetch Grades:', request.status, 'response:',this.response)
			var obj = JSON.parse(this.response)
			var payload = JSON.stringify(obj, undefined, 2);
			console.log('payload:', payload);
			gradeselement.innerHTML = payload;
			// table view
			var ge = '<hr /><table><tr><th>UserId</th><th>Name</th><th>Score</th></tr>';
			for (i = 0; i < obj.length; ++i) {
				ge += '<tr><td>' + obj[i].userId + '</td><td>' + getMemberName(obj[i].userId) + '</td><td>' + obj[i].resultScore + '</td></tr>';
			}
			ge += '</table>';
			gradelist.innerHTML = ge
	 	}
		request.send();
	}

	function getMemberName(userId) {
		console.log('get member name for: ' + userId);
		var member = memberMap[userId];
		console.log('member found: ' + JSON.stringify(member));
		if(member && member.name) {
			return member.name;
		} 
		return "";
	}

	</script>

		</head><body>
			<h1>tool</h1>
			<table border="1">
  			{{ range $key, $value := .Claims }}
           <tr><td><strong>{{ $key }}</strong></td><td>{{ $value }}</td></tr>
		  	{{ end }}
 			  <tr><td colspan="2">
				  Payload: <img width="300" src="{{.DoggoSrc}}" />
				</td></tr>
				<!-- MEMBER LOOKUP -->
				<tr><td><button onclick="fetchMembers()">Fetch Members</button></td>
					<td><pre id='memberelement'></pre></td></tr>
				<!-- GRADING -->
				<tr><td>Member Grading</td><td>
				    <span id='gradeamember' />
				  </td></tr>
				<tr><td>Last Grade Response</td>
					<td><pre id='scoreelement'></pre></td></tr>
				<!-- GRADE LOOKUP -->
				<tr><td><button onclick="fetchGrades()">Fetch Grades</button></td>
					<td>
						<pre id='gradeselement'></pre>
						<div id='gradelist'></div>

					</td></tr>
			</table>
			
	  </body></html>
	`
)

var (
	store                       sessions.Store
	cache                       ltiCache.Cache
	regDS                       registrationDatastore.RegistrationDatastore
	exampleOidcLogin            *lti.OidcLogin
	msgLaunchHandlerCreator     func(http.Handler) http.Handler
	nrpsGetMemberHandlerCreator func(http.Handler) http.Handler
	agsPutGradeHandlerCreator   func(http.Handler) http.Handler
	agsGetGradeHandlerCreator   func(http.Handler) http.Handler
	examplePayloadTemplate      *template.Template
	loggingHandler              http.Handler
)

func init() {
	// TODO: these keys should be stored somewhere and passed in via env
	// store = sessions.NewCookieStore(securecookie.GenerateRandomKey(32), securecookie.GenerateRandomKey(32))
	// use FileSystem store, since cookies can't hold the size of the jwt data
	// TODO: this should really be redis or something
	fsStoreRootPath := os.TempDir()
	fsStoreMaxLength := 1024 * 64
	fsStore := sessions.NewFilesystemStore(fsStoreRootPath, securecookie.GenerateRandomKey(32), securecookie.GenerateRandomKey(32))
	store = fsStore
	fsStore.MaxLength(fsStoreMaxLength)
	log.Printf("FsStore created, maxLen: %d, root dir: %s", fsStoreMaxLength, fsStoreRootPath)
	cache = ltiCache.NewSessionStoreCache(store, sessionCookieName)
	// TODO: this should be passed in via param
	regDS, err := registrationDatastore.NewJsonRegistrationDatastore("./registrationDatastore/registrations.json")
	if err != nil {
		panic("no registration datastore was found!")
	}
	exampleOidcLogin = lti.NewOidcLogin(regDS, cache, store, fmt.Sprintf("%s%s", getBaseURL(), exampleLaunchURL), sessionCookieName)

	examplePayloadTemplate, err = template.New("page").Parse(examplePayloadTemplateStr)
	if err != nil {
		panic("parse error in examplePayloadTemplate must be fixed!")
	}
	msgLaunchHandlerCreator = lti.MessageLaunchHandlerCreator(regDS, cache, store, sessionCookieName, debugFlag)
	nrpsGetMemberHandlerCreator = lti.NrpsGetMemberHandlerCreator(regDS, cache, store, sessionCookieName, debugFlag)
	exampleLineItem := &lti.LineItem{ScoreMax: 100, Label: "Example LI", Tag: "example_li"}
	agsPutGradeHandlerCreator = lti.AgsPutGradeHandlerCreator(regDS, cache, store, sessionCookieName, debugFlag, exampleLineItem)
	agsGetGradeHandlerCreator = lti.AgsGetGradesHandlerCreator(regDS, cache, store, sessionCookieName, debugFlag, exampleLineItem)
	loggingHandler = http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		log.Printf("done: /%s [launchID=%q]\n", req.URL.Path[1:], lti.GetLaunchID(req))
	})
}

func rootHandler(w http.ResponseWriter, req *http.Request) {
	fmt.Fprintf(w, "ok: %s\n", req.URL.Path[1:])
}

func otherHandler(w http.ResponseWriter, req *http.Request) {
	fmt.Fprintf(w, "other: %s\n", req.URL.Path[1:])
}

func examplePayloadHandler(w http.ResponseWriter, req *http.Request) {
	claims := lti.GetClaims(req)
	launchID := lti.GetLaunchID(req)
	log.Printf("examplePayloadHandler: launchID=%q", launchID)
	data := struct {
		Claims         jwt.MapClaims
		MemberPathPart string
		ScorePathPart  string
		GradesPathPart string
		LaunchID       string
		DoggoSrc       template.URL
	}{
		Claims:         claims,
		MemberPathPart: "members",
		ScorePathPart:  "grade",
		GradesPathPart: "grades",
		LaunchID:       launchID,
		DoggoSrc:       template.URL(doggoSrc),
	}

	if err := examplePayloadTemplate.Execute(w, data); err != nil {
		log.Printf("template failed to execute: %v", err)
	}
	log.Printf("done exampleLaunch %s", req.URL.String())
}

func main() {
	log.Println("Starting...")
	http.HandleFunc("/", rootHandler)
	http.Handle(exampleLoginURL, exampleOidcLogin.LoginRedirectHandler())
	http.Handle(exampleLaunchURL, msgLaunchHandlerCreator(http.HandlerFunc(examplePayloadHandler)))
	http.Handle(exampleMembersURL, nrpsGetMemberHandlerCreator(loggingHandler))
	http.Handle(exampleGradeURL, agsPutGradeHandlerCreator(loggingHandler))
	http.Handle(exampleGradesURL, agsGetGradeHandlerCreator(loggingHandler))
	// TODO: port should be a param
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", defaultPort), nil))
}

func getBaseURL() string {
	proto := "http"
	portStr := fmt.Sprintf(":%d", defaultPort)
	return fmt.Sprintf("%s://%s%s", proto, "localhost", portStr)
}
