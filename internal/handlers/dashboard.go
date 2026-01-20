package handlers

import (
	"GoBI/internal/database"
	"html/template"
	"net/http"
	"time"
)

type DashboardData struct {
	Stats        []Stat
	Results      []map[string]interface{}
	Columns      []TableColumn
	DatabaseName string
	Year         int
}

type Stat struct {
	Label string
	Value string
	Trend string
	Up    bool
}

var pool *database.CursorPool

func SetPool(p *database.CursorPool) {
	pool = p
}

func DashboardHandler(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.ParseFiles(
		"ui/templates/dashboard.html",
		"ui/templates/partials/nav.html",
		"ui/templates/partials/stats.html",
		"ui/templates/partials/table.html",
		"ui/templates/partials/footer.html",
	))

	data := DashboardData{
		Stats: []Stat{
			{Label: "Total Revenue", Value: "$124,500", Trend: "+12.5%", Up: true},
			{Label: "Active Users", Value: "1,240", Trend: "+5.2%", Up: true},
			{Label: "Bounce Rate", Value: "24.5%", Trend: "-2.1%", Up: false},
			{Label: "Server Load", Value: "14%", Trend: "Stable", Up: true},
		},
		DatabaseName: dbName,
		Year:         time.Now().Year(),
	}

	// Mock data for initial table
	data.Columns = []TableColumn{
		{Name: "ID", Label: "ID"},
		{Name: "Status", Label: "Status"},
		{Name: "Code", Label: "Code"},
		{Name: "Name", Label: "Name"},
		{Name: "Latest", Label: "Latest"},
	}
	data.Results = []map[string]interface{}{
		{"ID": "1001", "Status": "Active", "Code": "BIO-01", "Name": "System Alpha", "Latest": "2026-01-19"},
		{"ID": "1002", "Status": "Pending", "Code": "BIO-02", "Name": "System Beta", "Latest": "2026-01-20"},
	}

	tmpl.Execute(w, data)
}
