package main

import (
	"fmt"
	"html/template"
)

//tmpl help wirting the "pure string below"
func tmpl(page string) *template.Template { return template.Must(template.New("").Parse(page)) }

//Dashboard is the host object (and more)
// at least here is the struct definition, close to the template
type Dashboard struct {
	Title     string
	JobMatrix [][]Job
}

func (d *Dashboard) NameFontSize() string {
	//easier: number of row
	if len(d.JobMatrix) < 1 {
		return "12px" //does not matter
	}
	n := float64(len(d.JobMatrix))
	// full cell size would be 100/n vh
	size := 100 / n

	//let's make it half of the cell
	size /= 2
	return fmt.Sprintf("%vvh", size)

}

//VersionFont computes the version font as big as possible
func (d *Dashboard) VersionFontSize() string {
	//each version is 40 char width
	//there are n cols
	// there is then 40*n char to display
	// so it must be 100/40n vw

	// each char width is half the height,
	// so the font size is twice as musch
	if len(d.JobMatrix) < 1 {
		return "12px" //does not matter
	}
	n := float64(len(d.JobMatrix[0]))
	//maxsize (from width consideration)
	size := 4 / n

	//let's say that half the cell is prettier
	size /= 2

	return fmt.Sprintf("%vvw", size)
}

type Job struct {
	Name    string
	Status  string //css class for it's status
	Version string // a unique version (a sha1)
}

var dashboard = tmpl(`
<!DOCTYPE html>
	<html>
	<head>
	<meta http-equiv="refresh" content="10">
	<title>{{.Title}}</title>
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
		font-size:{{.NameFontSize}};
		text-align: center;
	}
	.version {
		font-size:{{.VersionFontSize}};

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
				<td class="{{.Status}}">

					<div>{{.Name}}</div>
					<div class="version">{{.Version}}</div>

				</td>
				{{end}}
			</tr>
			{{end}}

	</body></html>
	`)
