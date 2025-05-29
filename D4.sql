SELECT f.flight_no, tf.fare_conditions, tf.amount price
FROM ticket_flights tf
INNER JOIN flights f ON tf.flight_id = f.flight_id
GROUP BY f.flight_no, tf.fare_conditions, tf.amount;

