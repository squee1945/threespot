package pages

import "html/template"

type IndexArgs struct {
	PlayerID   string
	PlayerName string
}

var indexTemplateStr = `
<html>
<head><title>Kaiser</title></head>
<body>
<h1>Kaiser</h1>
<p>Welcome back {{.PlayerName}} (ID: {{.PlayerID}})</p>
<p>
  <form action='/setname' method='post'>
  Set your name: <input name='newname'>
  <br>
  <input type='submit'>
  </form>
</p>
</body>
</html>
`

var IndexPage = template.Must(template.New("index").Parse(indexTemplateStr))
