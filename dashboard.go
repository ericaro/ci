package ci

import (
	"html/template"
)

func tmpl(page string) *template.Template {
	return template.Must(template.New("").Parse(page))
}

//
var dashboard = tmpl(`
<!DOCTYPE html>
	<html>
	<head>
	<meta http-equiv="refresh" content="10" >
	<title>CI joe</title>
	<style>
	html,body,table{
		width:100%;
		height:100%;
		margin:0px;
		padding:0px;
		border-spacing: 0px;
	}
	td {
		font-family: "Courier New", Courier, monospace;
		font-size:5vh;
		text-align: center;
	}
	.small {
		font-size:1.3vw;

	}
	.running {
		color:  #F3F2D6;
		background-color:  #0C00F3;

	}
	.success {
		color:  #0C00F3;
		background-color:  #9FF8A5;
	}
	.failed {
		color:  #06052E;
		background-color:  #FF9C9C;
	}

</style>
	</head>
	<body>
		<table>
		{{range .JobMatrix}}
				<tr>
				{{range .}}
				<td class="{{.CssClass}}">
				<div>{{.Name}}</div>
				<div class="small">{{.Version}}</div>

				</td>{{end}}
				</tr>
		{{end}}
	</body></html>

	`)
