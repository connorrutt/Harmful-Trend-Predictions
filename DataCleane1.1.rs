use std::error::Error;
use rusqlite::{Connection, params};

fn main() -> Result<(), Box<dyn Error>> {
    let conn = Connection::open("ratings.db")?;

    conn.execute(
        "CREATE TABLE IF NOT EXISTS cleaned_ratings (
            user_id TEXT,
            item_id TEXT,
            rating REAL
        )",
        [],
    )?;

    conn.execute(
        "INSERT INTO cleaned_ratings
        SELECT DISTINCT user_id, item_id, rating
        FROM ratings",
        [],
    )?;

    conn.execute(
        "CREATE INDEX IF NOT EXISTS idx_user_id ON cleaned_ratings (user_id)",
        [],
    )?;
    conn.execute(
        "CREATE INDEX IF NOT EXISTS idx_item_id ON cleaned_ratings (item_id)",
        [],
    )?;

    conn.execute(
        "DELETE FROM cleaned_ratings
        WHERE rating < 3
           OR user_id IN (
                SELECT user_id
                FROM cleaned_ratings
                GROUP BY user_id
                HAVING COUNT(*) < 5
           )
           OR item_id IN (
                SELECT item_id
                FROM cleaned_ratings
                GROUP BY item_id
                HAVING COUNT(*) < 5
           )
           OR (user_id, item_id) IN (
                SELECT user_id, item_id
                FROM cleaned_ratings
                GROUP BY user_id, item_id
                HAVING COUNT(*) > 1
           )",
        [],
    )?;

    let count: i64 = conn.query_row(
        "SELECT COUNT(*) FROM cleaned_ratings",
        [],
        |row| row.get(0),
    )?;

    if count > 0 {
        conn.execute(
            "DELETE FROM cleaned_ratings
            WHERE rating > (
                    SELECT AVG(rating) + 3 * STDDEV(rating)
                    FROM cleaned_ratings
                ) OR rating < (
                    SELECT AVG(rating) - 3 * STDDEV(rating)
                    FROM cleaned_ratings
                )",
            [],
        )?;
    }

    conn.execute(
        "DELETE FROM cleaned_ratings
        WHERE user_id IN (
                SELECT user_id
                FROM cleaned_ratings
                GROUP BY user_id
                HAVING SUM(rating) < 10
           )
           OR item_id IN (
                SELECT item_id
                FROM cleaned_ratings
                GROUP BY item_id
                HAVING SUM(rating) < 10
           )",
        [],
    )?;


/// death is really easy to process
// when you just take it at face value. 
//Lifes life and I echo that. 
//But seeing others hurt almost in a way that is accepting grief just to know pat is in better place means a lot. 


    conn.execute(
        "CREATE TABLE IF NOT EXISTS final_ratings AS
        SELECT user_id, item_id, AVG(rating) AS average_rating
        FROM cleaned_ratings
        GROUP BY user_id, item_id",
        [],
    )?;

    Ok(())
}

-- Tried to clean up the rust a bit? unsure, need more data to scrape 
