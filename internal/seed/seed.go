package seed

import (
	"context"
	"fmt"
	"log"

	"github.com/BrunoMalagoli/bsmart-challenge/internal/auth"
	"github.com/BrunoMalagoli/bsmart-challenge/internal/db"
	"github.com/BrunoMalagoli/bsmart-challenge/internal/models"
)

func SeedAll(database *db.DB) error {
	ctx := context.Background()

	log.Println("Starting database seeding...")

	// Seed roles first (required for users)
	if err := SeedRoles(ctx, database); err != nil {
		return fmt.Errorf("failed to seed roles: %w", err)
	}

	if err := SeedUsers(ctx, database); err != nil {
		return fmt.Errorf("failed to seed users: %w", err)
	}

	// Seed categories (required for products)
	if err := SeedCategories(ctx, database); err != nil {
		return fmt.Errorf("failed to seed categories: %w", err)
	}

	if err := SeedProducts(ctx, database); err != nil {
		return fmt.Errorf("failed to seed products: %w", err)
	}

	log.Println("Database seeding completed successfully!")
	return nil
}

// SeedRoles creates admin and client roles
func SeedRoles(ctx context.Context, database *db.DB) error {
	log.Println("Seeding roles...")

	roles := []string{"admin", "client"}

	for _, roleName := range roles {
		// Check if role already exists
		_, err := database.GetRoleByName(ctx, roleName)
		if err == nil {
			log.Printf("  Role '%s' already exists, skipping", roleName)
			continue
		}

		// Create role
		role, err := database.CreateRole(ctx, roleName)
		if err != nil {
			return fmt.Errorf("failed to create role '%s': %w", roleName, err)
		}

		log.Printf("  Created role: %s (ID: %d)", role.Name, role.ID)
	}

	return nil
}

// SeedUsers creates test users
func SeedUsers(ctx context.Context, database *db.DB) error {
	log.Println("Seeding users...")

	// Get roles
	adminRole, err := database.GetRoleByName(ctx, "admin")
	if err != nil {
		return fmt.Errorf("failed to get admin role: %w", err)
	}

	clientRole, err := database.GetRoleByName(ctx, "client")
	if err != nil {
		return fmt.Errorf("failed to get client role: %w", err)
	}

	// Define test users
	users := []struct {
		email    string
		password string
		roleID   int
		roleName string
	}{
		{"admin@bsmart.com", "admin123", adminRole.ID, "admin"},
		{"client@bsmart.com", "client123", clientRole.ID, "client"},
		{"user1@bsmart.com", "password123", clientRole.ID, "client"},
		{"user2@bsmart.com", "password123", clientRole.ID, "client"},
	}

	for _, u := range users {
		// Check if user already exists
		exists, err := database.EmailExists(ctx, u.email)
		if err != nil {
			return fmt.Errorf("failed to check if user exists: %w", err)
		}

		if exists {
			log.Printf("  User '%s' already exists, skipping", u.email)
			continue
		}

		// Hash password
		passwordHash, err := auth.HashPassword(u.password)
		if err != nil {
			return fmt.Errorf("failed to hash password: %w", err)
		}

		user, err := database.CreateUser(ctx, u.email, passwordHash, &u.roleID)
		if err != nil {
			return fmt.Errorf("failed to create user: %w", err)
		}

		log.Printf("  Created user: %s (role: %s, ID: %d)", user.Email, u.roleName, user.ID)
	}

	return nil
}

func SeedCategories(ctx context.Context, database *db.DB) error {
	log.Println("Seeding categories...")

	categories := []struct {
		name        string
		description string
	}{
		{"Electronics", "Electronic devices and accessories"},
		{"Clothing", "Apparel and fashion items"},
		{"Books", "Books and publications"},
		{"Home & Garden", "Home decoration and gardening supplies"},
		{"Sports", "Sports equipment and accessories"},
		{"Toys", "Toys and games for all ages"},
		{"Food & Beverage", "Food products and beverages"},
		{"Health & Beauty", "Health and beauty products"},
	}

	for _, c := range categories {
		// Check if category already exists
		existingCategories, err := database.ListCategories(ctx, c.name)
		if err != nil {
			return fmt.Errorf("failed to check if category exists: %w", err)
		}

		if len(existingCategories) > 0 {
			log.Printf("  Category '%s' already exists, skipping", c.name)
			continue
		}

		// Create category
		desc := c.description
		category, err := database.CreateCategory(ctx, &models.CategoryCreateRequest{
			Name:        c.name,
			Description: &desc,
		})
		if err != nil {
			return fmt.Errorf("failed to create category: %w", err)
		}

		log.Printf("  Created category: %s (ID: %d)", category.Name, category.ID)
	}

	return nil
}

func SeedProducts(ctx context.Context, database *db.DB) error {
	log.Println("Seeding products...")

	// Get all categories
	allCategories, err := database.ListCategories(ctx, "")
	if err != nil {
		return fmt.Errorf("failed to get categories: %w", err)
	}

	if len(allCategories) == 0 {
		return fmt.Errorf("no categories found, cannot seed products")
	}

	// Create a map for easy category lookup
	categoryMap := make(map[string]int)
	for _, cat := range allCategories {
		categoryMap[cat.Name] = cat.ID
	}

	// Define sample products
	products := []struct {
		name        string
		description string
		price       float64
		stock       int
		categories  []string
	}{
		{"Laptop Dell XPS 13", "High-performance ultrabook with 11th Gen Intel Core i7", 1299.99, 15, []string{"Electronics"}},
		{"iPhone 14 Pro", "Latest Apple smartphone with A16 Bionic chip", 999.99, 25, []string{"Electronics"}},
		{"Wireless Mouse", "Ergonomic wireless mouse with USB receiver", 29.99, 100, []string{"Electronics"}},
		{"Men's T-Shirt", "Cotton crew neck t-shirt in various colors", 19.99, 200, []string{"Clothing"}},
		{"Women's Jeans", "Slim fit denim jeans", 49.99, 80, []string{"Clothing"}},
		{"Running Shoes", "Lightweight running shoes with cushioning", 89.99, 50, []string{"Clothing", "Sports"}},
		{"The Great Gatsby", "Classic American novel by F. Scott Fitzgerald", 12.99, 30, []string{"Books"}},
		{"Clean Code", "A Handbook of Agile Software Craftsmanship", 35.99, 40, []string{"Books"}},
		{"Garden Tools Set", "Complete set of essential gardening tools", 45.99, 25, []string{"Home & Garden"}},
		{"LED Desk Lamp", "Adjustable LED lamp with USB charging port", 34.99, 60, []string{"Electronics", "Home & Garden"}},
		{"Yoga Mat", "Non-slip exercise yoga mat", 24.99, 75, []string{"Sports"}},
		{"Tennis Racket", "Professional tennis racket for adults", 79.99, 20, []string{"Sports"}},
		{"LEGO Star Wars Set", "Building blocks set with 500 pieces", 59.99, 35, []string{"Toys"}},
		{"Board Game - Catan", "Strategy board game for 3-4 players", 44.99, 45, []string{"Toys"}},
		{"Organic Coffee Beans", "Premium arabica coffee beans, 1kg", 18.99, 120, []string{"Food & Beverage"}},
		{"Green Tea", "Organic green tea, 100 bags", 14.99, 90, []string{"Food & Beverage"}},
		{"Face Cream", "Anti-aging face cream with vitamin C", 29.99, 65, []string{"Health & Beauty"}},
		{"Shampoo & Conditioner Set", "Natural hair care set", 24.99, 80, []string{"Health & Beauty"}},
		{"Bluetooth Speaker", "Portable waterproof speaker with 12h battery", 49.99, 55, []string{"Electronics"}},
		{"Mechanical Keyboard", "RGB mechanical gaming keyboard", 119.99, 30, []string{"Electronics"}},
	}

	for _, p := range products {
		// Get category IDs
		var categoryIDs []int
		for _, catName := range p.categories {
			if catID, ok := categoryMap[catName]; ok {
				categoryIDs = append(categoryIDs, catID)
			}
		}

		if len(categoryIDs) == 0 {
			log.Printf("  Skipping product '%s' - no valid categories", p.name)
			continue
		}

		// Check if product already exists by name
		existingProducts, _, err := database.ListProducts(ctx, &db.PaginationParams{Page: 1, Limit: 100}, &db.FilterParams{Search: p.name})
		if err != nil {
			return fmt.Errorf("failed to check if product exists: %w", err)
		}

		// Simple check: if we find any products with the exact name, skip
		alreadyExists := false
		for _, existing := range existingProducts {
			if existing.Name == p.name {
				alreadyExists = true
				break
			}
		}

		if alreadyExists {
			log.Printf("  Product '%s' already exists, skipping", p.name)
			continue
		}

		desc := p.description
		product, err := database.CreateProduct(ctx, &models.ProductCreateRequest{
			Name:        p.name,
			Description: &desc,
			Price:       p.price,
			Stock:       p.stock,
			CategoryIDs: categoryIDs,
		})
		if err != nil {
			return fmt.Errorf("failed to create product '%s': %w", p.name, err)
		}

		log.Printf("  Created product: %s (ID: %d, Price: $%.2f, Stock: %d)", product.Name, product.ID, product.Price, product.Stock)
	}

	return nil
}
