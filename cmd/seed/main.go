package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"

	"auth-service/internal/db"
	"auth-service/internal/model"

	"github.com/joho/godotenv"
	"gorm.io/gorm"
)

func main() {
	godotenv.Load()

	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	database, err := db.Connect()
	if err != nil {
		log.Fatalf("db connect: %v", err)
	}

	switch os.Args[1] {

	case "batch":
		if len(os.Args) < 3 {
			log.Fatal("usage: seed batch <n>")
		}
		n, err := strconv.Atoi(os.Args[2])
		if err != nil {
			log.Fatalf("invalid count: %v", err)
		}
		seedBatch(database, n)

	case "root":
		flags := flag.NewFlagSet("root", flag.ExitOnError)
		email    := flags.String("email",    "root@dev.local", "admin email")
		password := flags.String("password", "root1234",       "admin password")
		flags.Parse(os.Args[2:])
		seedRootAdmin(database, *email, *password)

	case "all":
		seedAll(database)

	default:
		printUsage()
		os.Exit(1)
	}
}

// seedBatch creates n test users with the default user role.
func seedBatch(database *gorm.DB, count int) {
	roleID := mustRoleID(database, model.RoleUser)

	for i := 1; i <= count; i++ {
		user := model.User{
			Email:     fmt.Sprintf("user%d@dev.local", i),
			Password:  "password123",
			FirstName: fmt.Sprintf("User%d", i),
			LastName:  "Test",
			RoleID:    roleID,
		}
		if err := database.Where(model.User{Email: user.Email}).FirstOrCreate(&user).Error; err != nil {
			log.Printf("skip user%d: %v", i, err)
		}
	}

	fmt.Printf("seeded %d test users\n", count)
}

// seedRootAdmin creates an admin user with the given credentials.
// Defaults to root@dev.local / root1234 when called without flags.
func seedRootAdmin(database *gorm.DB, email, password string) {
	admin := model.User{
		Email:     email,
		Password:  password,
		FirstName: "Root",
		LastName:  "Admin",
		RoleID:    mustRoleID(database, model.RoleAdmin),
	}

	if err := database.Where(model.User{Email: admin.Email}).FirstOrCreate(&admin).Error; err != nil {
		log.Fatalf("seed root admin: %v", err)
	}

	fmt.Printf("seeded root admin  → %s / %s\n", email, password)
}

// seedAll creates one user per role — covers every case.
func seedAll(database *gorm.DB) {
	seedRootAdmin(database, "root@dev.local", "root1234")

	cases := []struct {
		role     model.UserRole
		email    string
		password string
	}{
		{model.RoleUser,  "user@dev.local",  "password123"},
		{model.RoleGuest, "guest@dev.local", "password123"},
	}

	for _, c := range cases {
		user := model.User{
			Email:     c.email,
			Password:  c.password,
			FirstName: string(c.role),
			LastName:  "Dev",
			RoleID:    mustRoleID(database, c.role),
		}
		if err := database.Where(model.User{Email: user.Email}).FirstOrCreate(&user).Error; err != nil {
			log.Printf("skip %s: %v", c.email, err)
			continue
		}
		fmt.Printf("seeded %-10s → %s / %s\n", c.role, c.email, c.password)
	}
}

// mustRoleID fetches a role ID by name or exits if not found.
func mustRoleID(database *gorm.DB, role model.UserRole) uint {
	var r model.Role
	if err := database.Where("name = ?", string(role)).First(&r).Error; err != nil {
		log.Fatalf("role %q not found — run migrations first", role)
	}
	return r.ID
}

func printUsage() {
	fmt.Println("usage:")
	fmt.Println("  seed batch <n>                        create n test users")
	fmt.Println("  seed root [-email=x] [-password=x]   create admin user")
	fmt.Println("  seed all                              one user per role")
}
