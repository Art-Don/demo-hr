package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"sort"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
)

type Reminder struct {
	ID           int    `json:"id"`
	Title        string `json:"title"`
	SubjectTitle string `json:"subject_title"`
	Content      string `json:"content"`
	Image        string `json:"image"`
	IsActive     bool   `json:"is_active"`
	IsChecked    bool   `json:"is_checked"`
	IsRemove     bool   `json:"is_remove"`
	DueDate      string `json:"due_date"`
	CreatedBy    string `json:"created_by"`
	UpdatedBy    string `json:"updated_by"`
	CreatedDate  string `json:"created_date"`
	UpdatedDate  string `json:"updated_date"`
}

const dataFilePath = "data.json"

var reminders []Reminder

type Req struct {
	ID        int    `json:"id"`
	Date      string `json:"date"`
	IsChecked bool   `json:"is_checked"`
}

func main() {
	app := fiber.New()

	app.Use(cors.New(cors.Config{
		// AllowOrigins:     "localhost",
		// AllowMethods:     strings.Join(allowMethod, ","),
		// AllowHeaders:     strings.Join(allowHeaders, ","),
		// AllowCredentials: false,
	}))

	app.Use(func(c *fiber.Ctx) error {
		log.Printf("Request: %s %s", c.Method(), c.OriginalURL())
		return c.Next()
	})

	app.Get("/reminder/list", func(c *fiber.Ctx) error {
		loadReminders()

		now := time.Now()

		var todayReminders []Reminder
		var futureReminders []Reminder
		var pastReminders []Reminder

		for _, reminder := range reminders {
			dueDate, err := time.Parse(time.RFC3339, reminder.DueDate)
			if err != nil {
				continue
			}

			if dueDate.Year() == now.Year() && dueDate.YearDay() == now.YearDay() {
				todayReminders = append(todayReminders, reminder)
			} else if dueDate.After(now) {
				futureReminders = append(futureReminders, reminder)
			} else {
				pastReminders = append(pastReminders, reminder)
			}
		}

		sort.Slice(todayReminders, func(i, j int) bool {
			dueDateI, errI := time.Parse(time.RFC3339, todayReminders[i].DueDate)
			dueDateJ, errJ := time.Parse(time.RFC3339, todayReminders[j].DueDate)

			if errI != nil || errJ != nil {
				return false
			}

			return dueDateI.Before(dueDateJ)
		})

		sort.Slice(futureReminders, func(i, j int) bool {
			dueDateI, errI := time.Parse(time.RFC3339, futureReminders[i].DueDate)
			dueDateJ, errJ := time.Parse(time.RFC3339, futureReminders[j].DueDate)

			if errI != nil || errJ != nil {
				return false
			}

			return dueDateI.Before(dueDateJ)
		})

		sort.Slice(pastReminders, func(i, j int) bool {
			dueDateI, errI := time.Parse(time.RFC3339, pastReminders[i].DueDate)
			dueDateJ, errJ := time.Parse(time.RFC3339, pastReminders[j].DueDate)

			if errI != nil || errJ != nil {
				return false
			}

			return dueDateI.After(dueDateJ)
		})

		allReminders := append(todayReminders, append(futureReminders, pastReminders...)...)

		sort.Slice(allReminders, func(i, j int) bool {
			return !allReminders[i].IsChecked && allReminders[j].IsChecked
		})

		return c.Status(http.StatusOK).JSON(fiber.Map{
			"code":    0,
			"message": "success",
			"result":  allReminders,
		})
	})

	app.Get("/reminder/id/:id", func(c *fiber.Ctx) error {
		loadReminders()
		id := c.Params("id")
		fmt.Println("Received ID:", id)

		var foundReminder *Reminder

		for _, reminder := range reminders {
			if fmt.Sprintf("%d", reminder.ID) == id {
				foundReminder = &reminder
				break
			}
		}

		if foundReminder != nil {
			fmt.Println("Found Reminder:", foundReminder)
			return c.JSON(fiber.Map{
				"code":    0,
				"message": "success",
				"result":  foundReminder,
			})
		}

		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"code":    1,
			"message": "Reminder not found",
		})
	})

	app.Post("/reminder/searchByDate", func(c *fiber.Ctx) error {
		loadReminders()
		var request struct {
			Date string `json:"date"` // Date in YYYY-MM format
		}

		if err := c.BodyParser(&request); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"code":    1,
				"message": "Invalid request body",
			})
		}

		var matchedReminders []Reminder
		for _, reminder := range reminders {
			parsedDate, err := time.Parse(time.RFC3339, reminder.DueDate)
			if err != nil {
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"code":    1,
					"message": "Error parsing reminder due date",
				})
			}

			reminderMonth := parsedDate.Format("2006-01")

			if reminderMonth == request.Date {
				matchedReminders = append(matchedReminders, reminder)
			}
		}

		// sort.Slice(matchedReminders, func(i, j int) bool {
		// 	return !matchedReminders[i].IsChecked && matchedReminders[j].IsChecked

		// })

		now := time.Now()

		var todayReminders []Reminder
		var futureReminders []Reminder
		var pastReminders []Reminder

		for _, reminder := range matchedReminders {
			dueDate, err := time.Parse(time.RFC3339, reminder.DueDate)
			if err != nil {
				continue
			}

			if dueDate.Year() == now.Year() && dueDate.YearDay() == now.YearDay() {
				todayReminders = append(todayReminders, reminder)
			} else if dueDate.After(now) {
				futureReminders = append(futureReminders, reminder)
			} else {
				pastReminders = append(pastReminders, reminder)
			}
		}

		sort.Slice(todayReminders, func(i, j int) bool {
			dueDateI, errI := time.Parse(time.RFC3339, todayReminders[i].DueDate)
			dueDateJ, errJ := time.Parse(time.RFC3339, todayReminders[j].DueDate)

			if errI != nil || errJ != nil {
				return false
			}

			return dueDateI.Before(dueDateJ)
		})

		sort.Slice(futureReminders, func(i, j int) bool {
			dueDateI, errI := time.Parse(time.RFC3339, futureReminders[i].DueDate)
			dueDateJ, errJ := time.Parse(time.RFC3339, futureReminders[j].DueDate)

			if errI != nil || errJ != nil {
				return false
			}

			return dueDateI.Before(dueDateJ)
		})

		sort.Slice(pastReminders, func(i, j int) bool {
			dueDateI, errI := time.Parse(time.RFC3339, pastReminders[i].DueDate)
			dueDateJ, errJ := time.Parse(time.RFC3339, pastReminders[j].DueDate)

			if errI != nil || errJ != nil {
				return false
			}

			return dueDateI.After(dueDateJ)
		})

		allReminders := append(todayReminders, append(futureReminders, pastReminders...)...)

		sort.Slice(allReminders, func(i, j int) bool {
			return !allReminders[i].IsChecked && allReminders[j].IsChecked
		})

		return c.JSON(fiber.Map{
			"code":    0,
			"message": "success",
			"result":  allReminders,
		})
	})

	app.Post("/reminder/update", func(c *fiber.Ctx) error {
		loadReminders()
		var request struct {
			ID        int  `json:"id"`
			IsChecked bool `json:"is_checked"`
		}

		if err := c.BodyParser(&request); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"code":    1,
				"message": "Invalid request body",
			})
		}
		fmt.Println("request :", request)

		var found bool

		for i, reminder := range reminders {
			if reminder.ID == request.ID {
				reminders[i].IsChecked = request.IsChecked
				found = true
				break
			}
		}

		if !found {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"code":    1,
				"message": "Reminder not found",
			})
		}

		if err := saveReminders(); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"code":    1,
				"message": "Error saving reminders to file",
			})
		}

		return c.JSON(fiber.Map{
			"code":    0,
			"message": "success",
		})
	})

	// Start the server
	app.Listen(":3000")
}

func loadReminders() error {
	file, err := os.Open(dataFilePath)
	if err != nil {
		return err
	}
	defer file.Close()

	bytes, err := ioutil.ReadAll(file)
	if err != nil {
		return err
	}

	err = json.Unmarshal(bytes, &reminders)
	if err != nil {
		return err
	}

	return nil
}

func saveReminders() error {
	bytes, err := json.MarshalIndent(reminders, "", "  ")
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(dataFilePath, bytes, 0644)
	if err != nil {
		return err
	}

	return nil
}
