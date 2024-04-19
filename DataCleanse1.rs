CREATE TABLE cleaned_ratings (
  user_id STRING,
  item_id STRING,
  rating FLOAT
);

INSERT INTO cleaned_ratings
SELECT DISTINCT user_id, item_id, rating
FROM ratings;

DELETE FROM cleaned_ratings
WHERE rating < 3;

DELETE FROM cleaned_ratings
WHERE user_id IN (
  SELECT user_id
  FROM cleaned_ratings
  GROUP BY user_id
  HAVING COUNT(*) < 5
);

DELETE FROM cleaned_ratings
WHERE item_id IN (
  SELECT item_id
  FROM cleaned_ratings
  GROUP BY item_id
  HAVING COUNT(*) < 5
);

DELETE FROM cleaned_ratings
WHERE (user_id, item_id) IN (
  SELECT user_id, item_id
  FROM cleaned_ratings
  GROUP BY user_id, item_id
  HAVING COUNT(*) > 1
);

DELETE FROM cleaned_ratings
WHERE (item_id, user_id) IN (
  SELECT item_id, user_id
  FROM cleaned_ratings
  GROUP BY item_id, user_id
  HAVING COUNT(*) > 1
);

DELETE FROM cleaned_ratings
WHERE rating > (
  SELECT AVG(rating) + 3 * STDDEV(rating)
  FROM cleaned_ratings
) OR rating < (
  SELECT AVG(rating) - 3 * STDDEV(rating)
  FROM cleaned_ratings
);

DELETE FROM cleaned_ratings
WHERE user_id IN (
  SELECT user_id
  FROM cleaned_ratings
  GROUP BY user_id
  HAVING SUM(rating) < 10
);

DELETE FROM cleaned_ratings
WHERE item_id IN (
  SELECT item_id
  FROM cleaned_ratings
  GROUP BY item_id
  HAVING SUM(rating) < 10
);

CREATE TABLE final_ratings AS
SELECT user_id, item_id, AVG(rating) AS average_rating
FROM cleaned_ratings
GROUP BY user_id, item_id
HAVING COUNT(*) > 1;

--Content assumes the data is stored with the metadata in a database in powerhouse. 
