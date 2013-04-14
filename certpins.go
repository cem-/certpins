//Copyright (C) 2013 Carl Mehner

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
<!DOCTYPE html>
<html>
  <head>
    <meta http-equiv="Content-Type" content="text/html; charset=UTF-8" />
  <title>Pin Creator</title>

  <style media="screen" type="text/css">
  html,
	
	#header {
		padding:10px;
	}
	#body {
		padding:10px;
		padding-bottom:2em;	/* Height of the footer */
	}
	#footer {
		position:absolute;
		bottom:0;
		height:2em;			/* Height of the footer */
	}
	#container {
		min-height:100%;
		position:relative;
	}
	p.output {
		padding-left:50px;
	}
  </style>
  </head>
  <body>
  <div id="container">
  <div id="header">
    <h1 style="font-family: 'Anaheim', sans-serif;">Certificate Pinning - Pin Creator</h1>
    <h5 style="padding-top:0em;margin-top:0em"><a href="/">home</a>&emsp;<a href="/about">about this tool</a></h5>
  </div>
  <div id="body">
`
//blob to write to bottom of every page
const foot = `
    </div>
    <br />
    <div id="footer">Copyright &copy; 2013 <a href="https://plus.google.com/115606090533118002214?rel=author">Carl Mehner</a></div>
    </div>
  </body>
</html>
`

//form to paste in your certificate
const certForm=`
<p>Drag and Drop a certificate file or paste in a certificate below.</p><p>Make sure the certificate is PEM, encoded with base64, and that they include the BEGIN and END headers</p>
    <form action="/pin" method="post">
      <div>
        <textarea name="cert" id="cert_drop" rows="20" cols="70" spellcheck="false" style="monospace;"></textarea>
<script>
function handleFileSelect(evt) {
    evt.stopPropagation();
    evt.preventDefault();

    var files = evt.dataTransfer.files; // FileList object.
    var reader = new FileReader();  
    reader.onload = function(event) {            
         document.getElementById('cert_drop').value = event.target.result;
    }        
    reader.readAsText(files[0],"UTF-8");
  }

  function handleDragOver(evt) {
    evt.stopPropagation();
    evt.preventDefault();
    evt.dataTransfer.dropEffect = 'copy'; // Explicitly show this is a copy.
  }

  // Setup the dnd listeners.
  var dropZone = document.getElementById('cert_drop');
  dropZone.addEventListener('dragover', handleDragOver, false);
  dropZone.addEventListener('drop', handleFileSelect, false);
</script>

      </div>
      <div><input type="submit" value="Create Cert Pin"></div>
    </form>
`

//handle about page
const aboutInfo=`
<p>Certificate pinning is a way to tell clients what cert or CA
they should be seeing when they connect to your website.<br />
Resources to find out more about pinning are included below:</p>
<li /><a href="https://tools.ietf.org/html/draft-ietf-websec-key-pinning">https://tools.ietf.org/html/draft-ietf-websec-key-pinning</a>
<li /><a href="http://www.imperialviolet.org/2011/05/04/pinning.html">http://www.imperialviolet.org/2011/05/04/pinning.html</a>
<li /><a href="http://www.thoughtcrime.org/blog/authenticity-is-broken-in-ssl-but-your-app-ha/">http://www.thoughtcrime.org/blog/authenticity-is-broken-in-ssl-but-your-app-ha/</a>
<p>This tool builds upon the go implementation of cert pinning in the key-pinning draft.</p>

<br />
<h2>Example Output</h2>
<h3>Certificate Pin created for www.google.com</h3><h4>To install in Chrome:</h4><p class="output">Go to <tt>chrome://net-internals/#hsts</tt> to manually add it to your chrome install</p><p class="output"><b>sha1/PH28jUH5JV3EJIKi5wVf64OTZK4=</b></p><h4>To use on your website:</h4><p class="output">Add to your <a href="https://en.wikipedia.org/wiki/HTTP_Strict_Transport_Security#Implementation">website's headers</a> using a method similar to HSTS (HTTP Strict Transport Security).<br />You <a href="https://tools.ietf.org/html/draft-ietf-websec-key-pinning#section-3.1">MUST</a> have a backup pin if you are using this method, otherwise, the browser will ignore the pin attempt.<br />You should get either a backup certificate, or the CA's cert that signs your server certs and run it again below<br />and add the line that starts with pin-sha256 into the headers you are sending to your clients.</p><p class="output"><b>Public-Key-Pins: max-age=31536000;<br />&emsp;pin-sha256="dhPKqMNshz05+vFbUc5C2HmcXReO04Fi+LtdzlydVD0=";</b></p><h4>To use in your Android app using Moxie's <a href="https://github.com/moxie0/AndroidPinning">AndroidPinning</a> library:</h4><p class="output">...<br />PinningTrustManager(new String[] {"<b>63f1353dd35fe084d20627b7861320d02aa18f14</b>"});<br />...</p><hr />
<p>
  <a rel="license" href="https://creativecommons.org/licenses/by-sa/3.0/deed.en_US">
    <img alt="Creative Commons License" style="border-width:0" src="https://i.creativecommons.org/l/by-sa/3.0/88x31.png" />
  </a>
  <br />
    <span xmlns:dct="http://purl.org/dc/terms/" property="dct:title">CertPins</span> by <a xmlns:cc="http://creativecommons.org/ns#" href="https://plus.google.com/115606090533118002214" rel="author" property="cc:attributionName" rel="cc:attributionURL">Carl Mehner</a> is licensed under a <a rel="license" href="https://creativecommons.org/licenses/by-sa/3.0/deed.en_US">Creative Commons Attribution-ShareAlike 3.0 Unported License</a>.<br />
    Based on a work at <a xmlns:dct="http://purl.org/dc/terms/" href="https://github.com/cem-/certpins" rel="dct:source">https://github.com/cem-/certpins</a>.
  </p>
`

//about page
func aboutHandler(w http.ResponseWriter, r *http.Request) {
  w.Header().Set("Strict-Transport-Security", "max-age=31536000")
  fmt.Fprint(w, head)
  fmt.Fprint(w, aboutInfo)
  fmt.Fprint(w, foot)
}

//main page
func handler(w http.ResponseWriter, r *http.Request) {
  w.Header().Set("Strict-Transport-Security", "max-age=31536000")
  fmt.Fprint(w, head)
  fmt.Fprint(w, certForm)
  fmt.Fprint(w, foot)
}

//pin page
func pinHandler(w http.ResponseWriter, r *http.Request) {
  //must parse before reading any data
  err := r.ParseForm()
  w.Header().Set("Strict-Transport-Security", "max-age=31536000")
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
      fmt.Fprint(w, "<h3>Certificate Pin created for " + cert.Subject.CommonName + "</h3><h4>To install in Chrome:</h4>")
      fmt.Fprint(w, "<p  class=\"output\">Go to <tt>chrome://net-internals/#hsts</tt> to manually add it to your chrome install</p>")
      fmt.Fprintf(w, "<p class=\"output\"><b>sha1/%s</b></p>", base64.StdEncoding.EncodeToString(digest))
      // print out in a form conforming to the key pinning draft
      fmt.Fprint(w, "<h4>To use on your website:</h4><p class=\"output\">Add to your <a href=\"https://en.wikipedia.org/wiki/HTTP_Strict_Transport_Security#Implementation\">website's headers</a> using a method similar to HSTS (HTTP Strict Transport Security).<br />You <a href=\"https://tools.ietf.org/html/draft-ietf-websec-key-pinning#section-3.1\">MUST</a> have a backup pin if you are using this method, otherwise, the browser will ignore the pin attempt.<br />You should get either a backup certificate, or the CA's cert that signs your server certs and run it again below<br />and add the line that starts with pin-sha256 into the headers you are sending to your clients.</p>")
      fmt.Fprintf(w, "<p class=\"output\"><b>Public-Key-Pins: max-age=31536000;<br />&emsp;pin-sha256=\"%s\";</b></p>", base64.StdEncoding.EncodeToString(digest2))
      fmt.Fprintf(w, "<h4>To use in your Android app using Moxie's <a href=\"https://github.com/moxie0/AndroidPinning\">AndroidPinning</a> library:</h4><p class=\"output\">...<br />PinningTrustManager(new String[] {\"<b>%x</b>\"});<br />...</p>", digest)
    }
  }

  fmt.Fprint(w, "<hr />")
  fmt.Fprint(w, certForm)
  fmt.Fprint(w, foot)
}

