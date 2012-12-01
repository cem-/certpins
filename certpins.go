//Copyright (C) 2012 Carl Mehner

//This program is free software: you can redistribute it and/or modify
//    it under the terms of the GNU General Public License as published by
//    the Free Software Foundation, either version 3 of the License, or
//    (at your option) any later version.
//
//    This program is distributed in the hope that it will be useful,
//    but WITHOUT ANY WARRANTY; without even the implied warranty of
//    MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
//    GNU General Public License for more details.
//
//    You should have received a copy of the GNU General Public License
//    along with this program.  If not, see <http://www.gnu.org/licenses/>.


package certpins

import (
          "crypto/sha1"
          "crypto/sha256"
          "crypto/x509"
          "encoding/base64"
          "encoding/pem"
          "fmt"
          "net/http"
)

func init() {
    http.HandleFunc("/about", aboutHandler)
    http.HandleFunc("/pin", pinHandler)
    http.HandleFunc("/", handler)
}

//blob to write to top of every page
const head =`
<html>
  <head><link href='http://fonts.googleapis.com/css?family=Anaheim' rel='stylesheet' type='text/css'>
  <style media="screen" type="text/css">
html,
	body {
		margin:0;
		padding:0;
		height:100%;
	}
	#container {
		min-height:100%;
		position:relative;
	}
	#header {
		padding:10px;

    max-width: 50em;
    margin-left: auto;
    margin-right: auto;
	}
	#body {
		padding:10px;
		padding-bottom:2em;	/* Height of the footer */
    max-width: 50em;
    margin-left: auto;
    margin-right: auto
	}
	#footer {
		position:absolute;
		bottom:0;
		width:100%;
		height:2em;			/* Height of the footer */

	}
  
  </style>
  </head>
  <body><div id="container">
  <div id="header">
    <h1 style="font-family: 'Anaheim', sans-serif;">Certificate Pinning - Pin Creator</h1>
    <h5 style="padding-top:0em;margin-top:0em"><a href="/">home</a>&emsp;<a href="/about">about this tool</a></h5>
    <br />
  </div>
  <div id="body">
`
//blob to write to bottom of every page
const foot = `
    </div>
    <div id="footer"><a rel="license" href="http://creativecommons.org/licenses/by-sa/3.0/deed.en_US"><img alt="Creative Commons License" style="border-width:0" src="http://i.creativecommons.org/l/by-sa/3.0/80x15.png" /></a> 2012 <a href="https://plus.google.com/115606090533118002214?rel=author">Carl Mehner</a></div>
    </div>
  </body>
</html>
`

//form to paste in your certificate
const certForm=`
    <p>Paste in a certificate below, PEM encoded with base64, including the BEGIN and END headers</p>
    <form action="/pin" method="post">
      <div><textarea name="cert" rows="20" cols="70" spellcheck="false" style="monospace"></textarea></div>
      <div><input type="submit" value="Create Cert Pin"></div>
    </form>
`

//handle about page
const aboutInfo=`
<p>Certificate pinning is a way to tell clients what cert or CA
they should be seeing when they connect to your website.<br />
Resources to find out more about pinning are included below:</p>
<li /><a href="http://tools.ietf.org/html/draft-ietf-websec-key-pinning">http://tools.ietf.org/html/draft-ietf-websec-key-pinning</a>
<li /><a href="http://www.imperialviolet.org/2011/05/04/pinning.html">http://www.imperialviolet.org/2011/05/04/pinning.html</a>
<li /><a href="http://www.thoughtcrime.org/blog/authenticity-is-broken-in-ssl-but-your-app-ha/">http://www.thoughtcrime.org/blog/authenticity-is-broken-in-ssl-but-your-app-ha/</a>
<p>This tool builds upon the go implementation of cert pinning in the key-pinning draft.</p>
<p>
  <a rel="license" href="http://creativecommons.org/licenses/by-sa/3.0/deed.en_US">
    <img alt="Creative Commons License" style="border-width:0" src="http://i.creativecommons.org/l/by-sa/3.0/88x31.png" />
  </a>
  <br />
    <span xmlns:dct="http://purl.org/dc/terms/" property="dct:title">CertPins</span> by <a xmlns:cc="http://creativecommons.org/ns#" href="https://plus.google.com/115606090533118002214" rel="author" property="cc:attributionName" rel="cc:attributionURL">Carl Mehner</a> is licensed under a <a rel="license" href="http://creativecommons.org/licenses/by-sa/3.0/deed.en_US">Creative Commons Attribution-ShareAlike 3.0 Unported License</a>.<br />
    Based on a work at <a xmlns:dct="http://purl.org/dc/terms/" href="https://github.com/cem-/certpins" rel="dct:source">https://github.com/cem-/certpins</a>.
  </p>
`

//about page
func aboutHandler(w http.ResponseWriter, r *http.Request) {
  fmt.Fprint(w, head)
  fmt.Fprint(w, aboutInfo)
  fmt.Fprint(w, foot)
}

//main page
func handler(w http.ResponseWriter, r *http.Request) {
  fmt.Fprint(w, head)
  fmt.Fprint(w, certForm)
  fmt.Fprint(w, foot)
}

//pin page
func pinHandler(w http.ResponseWriter, r *http.Request) {
  //must parse before reading any data
  err := r.ParseForm()
  fmt.Fprint(w, head)
  if err != nil {
      fmt.Fprint(w, "<p>Error Parsing Form Data: ")
      fmt.Fprint(w, err)
      fmt.Fprint(w, "</p>")
  }
  pemString := r.FormValue("cert")
  pemBytes := []byte(pemString)
  block, _ := pem.Decode(pemBytes)
  if block == nil {
    fmt.Fprint(w, "<p>no PEM structure found!</p>")
  } else {
    // get bytes from the cert
    derBytes := block.Bytes
    // put into a cert structure
    certs, err := x509.ParseCertificates(derBytes)
    if err != nil {
      fmt.Fprint(w, "<p>")
      fmt.Fprint(w, err)
      fmt.Fprint(w, "</p>")
    } else {
      // get first cert from the array
      cert := certs[0]
      // make sha1 and sha256 hashes for the sub pub key info
      h := sha1.New()
      h2 := sha256.New()
      h.Write(cert.RawSubjectPublicKeyInfo)
      h2.Write(cert.RawSubjectPublicKeyInfo)
      digest := h.Sum(nil)
      digest2 := h2.Sum(nil)
      // print out in a form for chrome to use
      fmt.Fprint(w, "<h3>Certificate Pin created for " + cert.Subject.CommonName + "</h3>")
      fmt.Fprint(w, "<p>Go to <tt>chrome://net-internals/#hsts</tt> to manually add it to your chrome install</p>")
      fmt.Fprintf(w, "<p><b>sha1/%s</b></p>", base64.StdEncoding.EncodeToString(digest))
      // print out in a form conforming to the key pinning draft
      fmt.Fprint(w, "<p>and/or add it to your <a href=\"http://en.wikipedia.org/wiki/HTTP_Strict_Transport_Security#Implementation\">website's headers</a> using a method similar to HSTS (HTTP Strict Transport Security).<br />You <a href=\"http://tools.ietf.org/html/draft-ietf-websec-key-pinning#section-3.1\">MUST</a> have a backup pin if you are using this method, otherwise, the browser will ignore the pin attempt.</p>")
      fmt.Fprintf(w, "<p><b>Public-Key-Pins: max-age=31536000;<br />&emsp;pin-sha1=\"%s\";<br />&emsp;pin-sha256=\"%s\";</b></p>", base64.StdEncoding.EncodeToString(digest), base64.StdEncoding.EncodeToString(digest2))
    }
  }

  fmt.Fprint(w, "<hr />")
  fmt.Fprint(w, certForm)
  fmt.Fprint(w, foot)
}

