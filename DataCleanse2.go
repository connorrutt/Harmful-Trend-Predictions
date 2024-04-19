package main

import (
	"database/sql"
	"fmt"

	_ "(Link to original DB)"
)

func main() {
	db, err := sql.Open("postgres", "user=myuser password=mypassword host=localhost dbname=mydatabase")
	if err != nil {
		fmt.Println("Error connecting to the database:", err)
		return
	}
	defer db.Close()

	_, err = db.Exec("CREATE TABLE cleaned_ratings (user_id text, item_id text, rating float)")
	if err != nil {
		fmt.Println("Error creating cleaned_ratings table:", err)
		return
	}

	_, err = db.Exec("INSERT INTO cleaned_ratings SELECT DISTINCT user_id, item_id, rating FROM ratings")
	if err != nil {
		fmt.Println("Error inserting data into cleaned_ratings table:", err)
		return
	}

	_, err = db.Exec("DELETE FROM cleaned_ratings WHERE rating < 3")
	if err != nil {
		fmt.Println("Error deleting low-quality ratings:", err)
		return
	}

	_, err = db.Exec("DELETE FROM cleaned_ratings WHERE user_id IN (SELECT user_id FROM cleaned_ratings GROUP BY user_id HAVING COUNT(*) < 5)")
	if err != nil {
		fmt.Println("Error deleting users with less than 5 ratings:", err)
		return
	}

	_, err = db.Exec("DELETE FROM cleaned_ratings WHERE item_id IN (SELECT item_id FROM cleaned_ratings GROUP BY item_id HAVING COUNT(*) < 5)")
	if err != nil {
		fmt.Println("Error deleting items with less than 5 ratings:", err)
		return
	}

	_, err = db.Exec("DELETE FROM cleaned_ratings WHERE (user_id, item_id) IN (SELECT user_id, item_id FROM cleaned_ratings GROUP BY user_id, item_id HAVING COUNT(*) > 1)")
	if err != nil {
		fmt.Println("Error deleting users who have rated the same item multiple times:", err)
		return
	}

	_, err = db.Exec("DELETE FROM cleaned_ratings WHERE (item_id, user_id) IN (SELECT item_id, user_id FROM cleaned_ratings GROUP BY item_id, user_id HAVING COUNT(*) > 1)")
	if err != nil {
		fmt.Println("Error deleting items that have been rated by the same user multiple times:", err)
		return
	}

	_, err = db.Exec("DELETE FROM cleaned_ratings WHERE rating > (SELECT AVG(rating) + 3 * STDDEV(rating) FROM cleaned_ratings) OR rating < (SELECT AVG(rating) - 3 * STDDEV(rating) FROM cleaned_ratings)")
	if err != nil {
		fmt.Println("Error deleting ratings that are more than 3 standard deviations away from the mean:", err)
		return
	}

	_, err = db.Exec("DELETE FROM cleaned_ratings WHERE user_id IN (SELECT user_id FROM cleaned_ratings GROUP BY user_id HAVING SUM(rating) < 10)")
	if err != nil {
		fmt.Println("Error deleting users and items with a total rating sum less than 10:", err)
		return
	}

	_, err = db.Exec("DELETE FROM cleaned_ratings WHERE item_id IN (SELECT item_id FROM cleaned_ratings GROUP BY item_id HAVING SUM(rating) < 10)")
	if err != nil {
		fmt.Println("Error deleting users and items with a total rating sum less than 10:", err)
		return
	}

	_, err = db.Exec("CREATE TABLE final_ratings AS SELECT user_id, item_id, AVG(rating) AS average_rating FROM cleaned_ratings GROUP BY user_id, item_id HAVING COUNT(*) > 1")
	if err != nil {
		fmt.Println("Error creating final_ratings table:", err)
		return
	}
} //content assumes the metadata is present in the powerhouse DB
