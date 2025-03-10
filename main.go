package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/joho/godotenv"
	_ "github.com/mattn/go-sqlite3"
)

// Styles
var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FAFAFA")).
			Background(lipgloss.Color("#7D56F4")).
			PaddingLeft(2).
			PaddingRight(2).
			MarginBottom(1)

	inputLabelStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#7D56F4")).
			Bold(true)

	infoStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFAD00"))

	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF0000"))

	resultHeaderStyle = lipgloss.NewStyle().
				Underline(true).
				Bold(true).
				Foreground(lipgloss.Color("#7D56F4")).
				PaddingTop(1).
				PaddingBottom(1)

	resultValueStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#2ECC71"))

	tableHeaderStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("#FFFFFF")).
				Background(lipgloss.Color("#333333")).
				PaddingLeft(1).
				PaddingRight(1)

	tableRowStyle = lipgloss.NewStyle().
			PaddingLeft(1).
			PaddingRight(1)

	tableRowAltStyle = lipgloss.NewStyle().
				Background(lipgloss.Color("#EEEEEE")).
				Foreground(lipgloss.Color("#333333")).
				PaddingLeft(1).
				PaddingRight(1)
)

// Weather data structures
type WeatherData struct {
	Location      string
	Date          time.Time
	MaxTemp       float64
	MinTemp       float64
	Humidity      float64
	Precipitation float64
}

type MonthlyAggregate struct {
	Month       time.Month
	AvgMaxTemp  float64
	AvgMinTemp  float64
	MaxTemp     float64
	MinTemp     float64
	AvgHumidity float64
	TotalPrecip float64
}

// Application state model
type model struct {
	inputs       []textinput.Model
	spinner      spinner.Model
	db           *sql.DB
	apiKey       string
	currentState string
	location     string
	errorMsg     string
	yearData     []WeatherData
	monthlyData  []MonthlyAggregate
	annualStats  struct {
		maxTemp     float64
		maxTempDate time.Time
		minTemp     float64
		minTempDate time.Time
		avgTemp     float64
		totalPrecip float64
	}
	loading  bool
	quitting bool
}

func initialModel() model {
	// Load environment variables
	godotenv.Load()

	// Set up text inputs
	locationInput := textinput.New()
	locationInput.Placeholder = "City, State"
	locationInput.Focus()
	locationInput.Width = 30
	locationInput.Prompt = inputLabelStyle.Render("Location: ")

	// Create spinner
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))

	return model{
		inputs:       []textinput.Model{locationInput},
		spinner:      s,
		currentState: "input",
		apiKey:       os.Getenv("WEATHER_API_KEY"),
		loading:      false,
		quitting:     false,
	}
}

func (m model) Init() tea.Cmd {
	return tea.Batch(textinput.Blink, m.spinner.Tick)
}

// Update model based on messages
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			// Clean up db connection
			if m.db != nil {
				m.db.Close()
			}
			m.quitting = true
			return m, tea.Quit

		case "enter":
			if m.currentState == "input" {
				m.location = m.inputs[0].Value()
				if m.location == "" {
					m.errorMsg = "Please enter a location"
					return m, nil
				}

				m.loading = true
				m.errorMsg = ""

				// Connect to database
				return m, tea.Batch(
					m.connectToDatabase,
					m.spinner.Tick,
				)
			} else if m.currentState == "results" {
				// Return to input state
				m.currentState = "input"
				m.inputs[0].SetValue("")
				m.inputs[0].Focus()
				return m, nil
			}
		}

	case errMsg:
		m.errorMsg = msg.Error()
		m.loading = false
		m.currentState = "input"
		return m, nil

	case dbConnectedMsg:
		// Database connected, fetch weather data
		return m, m.fetchWeatherData

	case weatherDataFetchedMsg:
		// Weather data has been fetched and saved to DB
		return m, m.analyzeWeatherData

	case weatherDataAnalyzedMsg:
		// Analysis complete, show results
		m.loading = false
		m.currentState = "results"
		return m, nil

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	}

	// Handle text input updates
	if m.currentState == "input" {
		var cmd tea.Cmd
		m.inputs[0], cmd = m.inputs[0].Update(msg)
		return m, cmd
	}

	return m, nil
}

func (m model) View() string {
	if m.quitting {
		return "Thank you for using Weather Trend Analyzer!\n"
	}

	s := titleStyle.Render(" Weather Trend Analyzer ") + "\n\n"

	switch m.currentState {
	case "input":
		s += m.inputs[0].View() + "\n\n"
		s += "Press Enter to analyze weather trends\n"
		s += "Press Ctrl+C to quit\n\n"

		if m.loading {
			s += m.spinner.View() + " Loading weather data...\n"
		}

		if m.errorMsg != "" {
			s += errorStyle.Render("Error: "+m.errorMsg) + "\n"
		}

	case "results":
		// Display location
		s += fmt.Sprintf("Weather analysis for: %s\n\n", m.location)

		// Annual statistics
		s += resultHeaderStyle.Render("Annual Statistics") + "\n"
		s += fmt.Sprintf("Highest Temperature: %s (on %s)\n",
			resultValueStyle.Render(fmt.Sprintf("%.1f°C", m.annualStats.maxTemp)),
			m.annualStats.maxTempDate.Format("Jan 2"))
		s += fmt.Sprintf("Lowest Temperature: %s (on %s)\n",
			resultValueStyle.Render(fmt.Sprintf("%.1f°C", m.annualStats.minTemp)),
			m.annualStats.minTempDate.Format("Jan 2"))
		s += fmt.Sprintf("Average Temperature: %s\n",
			resultValueStyle.Render(fmt.Sprintf("%.1f°C", m.annualStats.avgTemp)))
		s += fmt.Sprintf("Total Precipitation: %s\n\n",
			resultValueStyle.Render(fmt.Sprintf("%.1f mm", m.annualStats.totalPrecip)))

		// Monthly breakdown
		s += resultHeaderStyle.Render("Monthly Breakdown") + "\n"

		// Table header
		headers := []string{"Month", "Avg High", "Avg Low", "Max", "Min", "Humidity", "Precip"}
		headerRow := ""
		for _, h := range headers {
			headerRow += tableHeaderStyle.Render(fmt.Sprintf("%-10s", h))
		}
		s += headerRow + "\n"

		// Table data
		for i, m := range m.monthlyData {
			rowStyle := tableRowStyle
			if i%2 == 1 {
				rowStyle = tableRowAltStyle
			}

			row := rowStyle.Render(fmt.Sprintf("%-10s", m.Month.String()))
			row += rowStyle.Render(fmt.Sprintf("%-10.1f", m.AvgMaxTemp))
			row += rowStyle.Render(fmt.Sprintf("%-10.1f", m.AvgMinTemp))
			row += rowStyle.Render(fmt.Sprintf("%-10.1f", m.MaxTemp))
			row += rowStyle.Render(fmt.Sprintf("%-10.1f", m.MinTemp))
			row += rowStyle.Render(fmt.Sprintf("%-10.1f", m.AvgHumidity))
			row += rowStyle.Render(fmt.Sprintf("%-10.1f", m.TotalPrecip))

			s += row + "\n"
		}

		s += "\nPress Enter to analyze another location\n"
		s += "Press Q to quit\n"
	}

	return s
}

// Custom message types
type errMsg struct{ err error }

func (e errMsg) Error() string { return e.err.Error() }

type dbConnectedMsg struct{}
type weatherDataFetchedMsg struct{}
type weatherDataAnalyzedMsg struct{}

// Connect to SQLite database
func (m model) connectToDatabase() tea.Msg {
	// Use a file in the current directory for SQLite
	dbPath := "./weather.db"

	// Connect to the database
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return errMsg{err: fmt.Errorf("unable to connect to database: %v", err)}
	}

	// Check connection
	if err := db.Ping(); err != nil {
		return errMsg{err: fmt.Errorf("unable to ping database: %v", err)}
	}

	// Create tables if they don't exist
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS weather_data (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			location TEXT NOT NULL,
			date TEXT NOT NULL, 
			max_temp REAL NOT NULL,
			min_temp REAL NOT NULL,
			humidity REAL NOT NULL,
			precipitation REAL NOT NULL,
			UNIQUE(location, date)
		)
	`)

	if err != nil {
		return errMsg{err: fmt.Errorf("unable to create tables: %v", err)}
	}

	m.db = db
	return dbConnectedMsg{}
}

// Fetch weather data from API or database
func (m model) fetchWeatherData() tea.Msg {
	// Check if we already have data for this location and year
	currentYear := time.Now().Year()
	startDate := time.Date(currentYear, 1, 1, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(currentYear, 12, 31, 0, 0, 0, 0, time.UTC)

	// Format dates for SQLite (which stores dates as text)
	startDateStr := startDate.Format("2006-01-02")
	endDateStr := endDate.Format("2006-01-02")

	// Format location for query
	location := strings.TrimSpace(m.location)

	// Check if we have data in database
	var count int
	err := m.db.QueryRow(`
		SELECT COUNT(*) FROM weather_data 
		WHERE location = ? AND date BETWEEN ? AND ?
	`, location, startDateStr, endDateStr).Scan(&count)

	if err != nil {
		return errMsg{err: fmt.Errorf("database query error: %v", err)}
	}

	// If we don't have enough data, fetch from API
	if count < 300 { // We want most days of the year
		// Generate and insert mock data for demonstration
		err = generateMockData(m.db, location, currentYear)
		if err != nil {
			return errMsg{err: fmt.Errorf("error generating mock data: %v", err)}
		}
	}

	// Now fetch the data from our database
	rows, err := m.db.Query(`
		SELECT location, date, max_temp, min_temp, humidity, precipitation 
		FROM weather_data 
		WHERE location = ? AND date BETWEEN ? AND ?
		ORDER BY date
	`, location, startDateStr, endDateStr)

	if err != nil {
		return errMsg{err: fmt.Errorf("error querying weather data: %v", err)}
	}
	defer rows.Close()

	// Parse rows into weather data
	var weatherData []WeatherData
	for rows.Next() {
		var wd WeatherData
		var dateStr string

		if err := rows.Scan(&wd.Location, &dateStr, &wd.MaxTemp, &wd.MinTemp, &wd.Humidity, &wd.Precipitation); err != nil {
			return errMsg{err: fmt.Errorf("error scanning row: %v", err)}
		}

		// Parse the date string from SQLite
		wd.Date, err = time.Parse("2006-01-02", dateStr)
		if err != nil {
			return errMsg{err: fmt.Errorf("error parsing date: %v", err)}
		}

		weatherData = append(weatherData, wd)
	}

	if err := rows.Err(); err != nil {
		return errMsg{err: fmt.Errorf("error iterating rows: %v", err)}
	}

	if len(weatherData) == 0 {
		return errMsg{err: fmt.Errorf("no weather data found for %s", m.location)}
	}

	m.yearData = weatherData
	return weatherDataFetchedMsg{}
}

// Analyze weather data
func (m model) analyzeWeatherData() tea.Msg {
	if len(m.yearData) == 0 {
		return errMsg{err: fmt.Errorf("no weather data to analyze")}
	}

	// Initialize variables for annual statistics
	m.annualStats.maxTemp = -100
	m.annualStats.minTemp = 100
	var sumTemp float64
	var totalDays int

	// Initialize monthly aggregates
	monthlyMap := make(map[time.Month]*MonthlyAggregate)
	for i := time.January; i <= time.December; i++ {
		monthlyMap[i] = &MonthlyAggregate{
			Month:   i,
			MinTemp: 100,  // Initialize to high value
			MaxTemp: -100, // Initialize to low value
		}
	}

	// Process each data point
	for _, wd := range m.yearData {
		// Annual stats
		if wd.MaxTemp > m.annualStats.maxTemp {
			m.annualStats.maxTemp = wd.MaxTemp
			m.annualStats.maxTempDate = wd.Date
		}
		if wd.MinTemp < m.annualStats.minTemp {
			m.annualStats.minTemp = wd.MinTemp
			m.annualStats.minTempDate = wd.Date
		}

		sumTemp += (wd.MaxTemp + wd.MinTemp) / 2
		totalDays++
		m.annualStats.totalPrecip += wd.Precipitation

		// Monthly stats
		month := wd.Date.Month()
		ma := monthlyMap[month]

		// Update max/min temperatures
		if wd.MaxTemp > ma.MaxTemp {
			ma.MaxTemp = wd.MaxTemp
		}
		if wd.MinTemp < ma.MinTemp {
			ma.MinTemp = wd.MinTemp
		}

		// Accumulate for averages
		ma.AvgMaxTemp += wd.MaxTemp
		ma.AvgMinTemp += wd.MinTemp
		ma.AvgHumidity += wd.Humidity
		ma.TotalPrecip += wd.Precipitation
	}

	// Calculate averages for monthly data
	m.monthlyData = []MonthlyAggregate{}
	for month := time.January; month <= time.December; month++ {
		ma := monthlyMap[month]

		// Count days in this month
		var daysInMonth int
		for _, wd := range m.yearData {
			if wd.Date.Month() == month {
				daysInMonth++
			}
		}

		// Only include months with data
		if daysInMonth > 0 {
			ma.AvgMaxTemp /= float64(daysInMonth)
			ma.AvgMinTemp /= float64(daysInMonth)
			ma.AvgHumidity /= float64(daysInMonth)
			m.monthlyData = append(m.monthlyData, *ma)
		}
	}

	// Calculate annual average temperature
	if totalDays > 0 {
		m.annualStats.avgTemp = sumTemp / float64(totalDays)
	}

	return weatherDataAnalyzedMsg{}
}

// Mock data generator for demonstration purposes
func generateMockData(db *sql.DB, location string, year int) error {
	// Start a transaction
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Prepare the insert statement
	stmt, err := tx.Prepare(`
		INSERT INTO weather_data (location, date, max_temp, min_temp, humidity, precipitation)
		VALUES (?, ?, ?, ?, ?, ?)
		ON CONFLICT(location, date) DO NOTHING
	`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	// Generate data for each day of the year
	for month := 1; month <= 12; month++ {
		// Get days in month
		daysInMonth := 31
		if month == 4 || month == 6 || month == 9 || month == 11 {
			daysInMonth = 30
		} else if month == 2 {
			// Simple leap year check
			if year%4 == 0 && (year%100 != 0 || year%400 == 0) {
				daysInMonth = 29
			} else {
				daysInMonth = 28
			}
		}

		for day := 1; day <= daysInMonth; day++ {
			date := time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC)
			dateStr := date.Format("2006-01-02")

			// Skip future dates
			if date.After(time.Now()) {
				continue
			}

			// Generate realistic seasonal variations based on month
			seasonalFactor := float64(month-1) / 11.0 // 0 to 1 through the year

			// Northern hemisphere seasons (adjust for southern hemisphere)
			var baseTemp float64
			if month >= 6 && month <= 8 { // Summer (Northern)
				baseTemp = 25
			} else if month >= 3 && month <= 5 { // Spring
				baseTemp = 15
			} else if month >= 9 && month <= 11 { // Fall
				baseTemp = 15
			} else { // Winter
				baseTemp = 5
			}

			// Add some randomness
			randomRange := 8.0
			maxTemp := baseTemp + ((rand.Float64() * randomRange) - (randomRange / 2))
			minTemp := maxTemp - 5 - (rand.Float64() * 5) // 5-10 degrees cooler at night

			// Humidity tends to be higher in summer
			humidity := 50.0 + ((rand.Float64() * 40) - 20)
			if humidity < 10 {
				humidity = 10
			} else if humidity > 100 {
				humidity = 100
			}

			// Precipitation varies by season
			var basePrecip float64
			if month >= 6 && month <= 8 { // Summer often has thunderstorms
				basePrecip = 3
			} else if month >= 3 && month <= 5 { // Spring
				basePrecip = 5
			} else if month >= 9 && month <= 11 { // Fall
				basePrecip = 4
			} else { // Winter
				basePrecip = 2
			}

			// Most days are dry, some days have rain
			precipitation := 0.0
			if rand.Float64() < 0.3 { // 30% chance of precipitation
				precipitation = basePrecip + (rand.Float64() * 10)
			}

			// Insert data
			_, err := stmt.Exec(location, dateStr, maxTemp, minTemp, humidity, precipitation)
			if err != nil {
				return err
			}
		}
	}

	return tx.Commit()
}

// Run the application
func main() {
	// Seed the random number generator
	rand.Seed(time.Now().UnixNano())

	p := tea.NewProgram(initialModel())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error running program: %v\n", err)
		os.Exit(1)
	}
}
