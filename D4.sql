CREATE TABLE flight_new_fare_diff AS
SELECT DISTINCT f.flight_no, f.aircraft_code, tf.fare_conditions, MAX(tf.amount) new_fare, MIN(tf.amount) old
FROM ticket_flights tf
INNER JOIN flights f ON tf.flight_id = f.flight_id
GROUP BY f.flight_no, f.aircraft_code, tf.fare_conditions
HAVING MAX(tf.amount) <> MIN(tf.amount)
ORDER BY f.aircraft_code;


ALTER TABLE bookings.seats
DROP CONSTRAINT seats_fare_conditions_check;

ALTER TABLE bookings.seats
ADD CONSTRAINT seats_fare_conditions_check
CHECK (((fare_conditions)::text = ANY (ARRAY[
        ('Economy'::character varying)::text, 
        ('EconomySec'::character varying)::text,
        ('Comfort'::character varying)::text, 
        ('Business'::character varying)::text
])));


ALTER TABLE ticket_flights
DROP CONSTRAINT ticket_flights_fare_conditions_check;

ALTER TABLE ticket_flights
ADD CONSTRAINT ticket_flights_fare_conditions_check
CHECK (((fare_conditions)::text = ANY (ARRAY[
        ('Economy'::character varying)::text, 
        ('EconomySec'::character varying)::text,
        ('Comfort'::character varying)::text, 
        ('Business'::character varying)::text
])));


UPDATE ticket_flights AS tf
SET fare_conditions = 'EconomySec'
FROM flights f
WHERE f.flight_id = tf.flight_id AND
EXISTS (SELECT 1 FROM flight_new_fare_diff
        WHERE flight_no = f.flight_no AND fare_conditions = tf.fare_conditions AND new_fare = tf.amount);


UPDATE seats AS s
SET fare_conditions = 'EconomySec'
FROM (
    SELECT DISTINCT f.aircraft_code, bp.seat_no
    FROM flights f
    INNER JOIN ticket_flights tf ON f.flight_id = tf.flight_id
    INNER JOIN boarding_passes bp ON tf.flight_id = bp.flight_id AND tf.ticket_no = bp.ticket_no
    WHERE tf.fare_conditions = 'EconomySec'
) tf
WHERE s.aircraft_code = tf.aircraft_code AND s.seat_no = tf.seat_no;


CREATE TABLE delivery_prices AS
SELECT DISTINCT f.aircraft_code, f.departure_airport, f.arrival_airport, tf.fare_conditions, tf.amount price
FROM flights f
INNER JOIN ticket_flights tf ON tf.flight_id = f.flight_id;
