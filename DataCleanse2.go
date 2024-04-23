package main

import (
    "database/sql"
    "fmt"
    "log"
    "os"

    _ "github.com/lib/pq"
)

func main() {
    user := os.Getenv("DB_USER")
    password := os.Getenv("DB_PASSWORD")
    host := os.Getenv("DB_HOST")
    dbname := os.Getenv("DB_NAME")

    connStr := fmt.Sprintf("user=%s password=%s host=%s dbname=%s", user, password, host, dbname)

    db, err := sql.Open("postgres", connStr)
    if err != nil {
        log.Fatal("Error connecting to the database:", err)
    }
    defer db.Close()

    err = db.Ping()
    if err != nil {
        log.Fatal("Error pinging the database:", err)
    }

    _, err = db.Exec(`
        CREATE TABLE IF NOT EXISTS cleaned_ratings (
            user_id text,
            item_id text,
            rating float
        )
    `)
    if err != nil {
        log.Fatal("Error creating cleaned_ratings table:", err)
    }

    tx, err := db.Begin()
    if err != nil {
        log.Fatal("Error starting a transaction:", err)
    }

    _, err = tx.Exec(`
        WITH cte AS (
            SELECT DISTINCT user_id, item_id, rating
            FROM ratings
            WHERE rating >= 3
        ),
        user_counts AS (
            SELECT user_id, COUNT(*) AS count
            FROM cte
            GROUP BY user_id
            HAVING COUNT(*) >= 5
        ),
        item_counts AS (
            SELECT item_id, COUNT(*) AS count
            FROM cte
            GROUP BY item_id
            HAVING COUNT(*) >= 5
        ),
        duplicate_ratings AS (
            SELECT user_id, item_id, COUNT(*) AS count
            FROM cte
            GROUP BY user_id, item_id
            HAVING COUNT(*) > 1
        ),
        outlier_ratings AS (
            SELECT *
            FROM cte
            WHERE rating > (SELECT AVG(rating) + 3 * STDDEV(rating) FROM cte)
               OR rating < (SELECT AVG(rating) - 3 * STDDEV(rating) FROM cte)
        ),
        low_sum_users AS (
            SELECT user_id, SUM(rating) AS sum
            FROM cte
            GROUP BY user_id
            HAVING SUM(rating) < 10
        ),
        low_sum_items AS (
            SELECT item_id, SUM(rating) AS sum
            FROM cte
            GROUP BY item_id
            HAVING SUM(rating) < 10
        )
        INSERT INTO cleaned_ratings
        SELECT cte.user_id, cte.item_id, cte.rating
        FROM cte
        INNER JOIN user_counts ON cte.user_id = user_counts.user_id
        INNER JOIN item_counts ON cte.item_id = item_counts.item_id
        LEFT JOIN duplicate_ratings ON cte.user_id = duplicate_ratings.user_id AND cte.item_id = duplicate_ratings.item_id
        LEFT JOIN outlier_ratings ON cte.user_id = outlier_ratings.user_id AND cte.item_id = outlier_ratings.item_id AND cte.rating = outlier_ratings.rating
        LEFT JOIN low_sum_users ON cte.user_id = low_sum_users.user_id
        LEFT JOIN low_sum_items ON cte.item_id = low_sum_items.item_id
        WHERE duplicate_ratings.user_id IS NULL
          AND duplicate_ratings.item_id IS NULL
          AND outlier_ratings.user_id IS NULL
          AND outlier_ratings.item_id IS NULL
          AND outlier_ratings.rating IS NULL
          AND low_sum_users.user_id IS NULL
          AND low_sum_items.item_id IS NULL
    `)
    if err != nil {
        tx.Rollback()
        log.Fatal("Error cleaning and filtering the data:", err)
    }

    _, err = tx.Exec(`
        CREATE TABLE final_ratings AS
        SELECT user_id, item_id, AVG(rating) AS average_rating
        FROM cleaned_ratings
        GROUP BY user_id, item_id
        HAVING COUNT(*) > 1
    `)
    if err != nil {
        tx.Rollback()
        log.Fatal("Error creating final_ratings table:", err)
    }

    err = tx.Commit()
    if err != nil {
        log.Fatal("Error committing the transaction:", err)
    }

    fmt.Println("Data cleaning and table creation completed successfully.")
}//content assumes the metadata is present in the powerhouse DB
