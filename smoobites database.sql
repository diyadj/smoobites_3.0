drop schema if exists smoobites;
create schema smoobites;
use smoobites;

CREATE TABLE users (
    id INT AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    email VARCHAR(100) UNIQUE NOT NULL,
    password VARCHAR(255) NOT NULL,
    role VARCHAR(50) NOT NULL,
    school VARCHAR(50)
);
insert into users (name, email, password, role,school)
values 
("Admin", "smoobites@gmail.com", "$2a$10$g6C1dZ2Sa/MCye.pJW55J.OoIcg9bCwa.71jm7ZqT4WAJWuWfhD3S", "admin", "NA"),
("Subway", "subway@smu.com", "$2a$10$IY/OUgDJQiLqlCb55RrSSeweHRpcyz7ki3PGiVD9uF2KWVnebqOBe", "vendor","SCIS1"),
("Providence", "providence@smu.com","$2a$10$YNuyAXclFzsAzcjaH4x7PenSANguNd9sFoD2DEJwUu3Leuvhpgm4K","vendor","SCIS1"),
("Nasi Lemak Ayam Taliwang","nasilemak@smu.com","$2a$10$6Ju4dYf/aLugQXilRgPEIuljDMGc.ZvwhNiSNSQ5GSLj0Cgu7ZY6a","vendor", "SOE/SCIS2");
-- create table for password resets
CREATE TABLE password_resets (
    email VARCHAR(100) UNIQUE NOT NULL,
    token VARCHAR(255) NOT NULL,
    expires_at DATETIME NOT NULL
);

CREATE TABLE food_items (
    id INT AUTO_INCREMENT PRIMARY KEY,
    food_name VARCHAR(255) NOT NULL,
    description TEXT,
    price VARCHAR(50) NOT NULL,
    prep_time VARCHAR(50) NOT NULL,
    image_path VARCHAR(255),
    vendor_id INT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (vendor_id) REFERENCES users(id) ON DELETE CASCADE
);
insert into food_items (food_name, description,price,prep_time,image_path,vendor_id)
values
    ('6 Inch Italian B.M.T',"An old-world favorite. Sliced beef salami, beef pepperoni and chicken ham and your choice of fresh vegetables and condiments served on freshly baked bread. Some say B.M.T. stands for biggest, meatiest, tastiest. We wouldn't disagree.. ", '10.20','2','uploads\italian.png',2),
    ('6 Inch Bulgogi Chicken', "Tender chicken marinated in the classic sweet and savoury Korean bulgogi sauce.",'8.40', '2', 'uploads\bulgogi.png',2),
    ('6 Inch Egg Mayo', "Our delicious concoction of hard boiled eggs and Mayonnaise. It's a local favorite for many Singaporeans and even satisfies some of our vegetarian customers.",'6.90', '2', 'uploads\egg.png',2),
    ('6 Inch Veggie Patty',"If you're a vegetarian or even a carnivore with a taste for green, you have to try this tasty vegetarian selection. The patty is made from vegetables and brown rice and its own unique spices.It's one of our favorites.", '9.10', '2', 'uploads\veggie.png',2),
    ('12 Inch Italian B.M.T',"An old-world favorite. Sliced beef salami, beef pepperoni and chicken ham and your choice of fresh vegetables and condiments served on freshly baked bread. Some say B.M.T. stands for biggest, meatiest, tastiest. We wouldn't disagree.. ", '15.70','3','uploads\italian.png',2),
    ('12 Inch Bulgogi Chicken', "Tender chicken marinated in the classic sweet and savoury Korean bulgogi sauce.",'13.90', '3', 'uploads\bulgogi.png',2),
    ('12 Inch Egg Mayo', "Our delicious concoction of hard boiled eggs and Mayonnaise. It's a local favorite for many Singaporeans and even satisfies some of our vegetarian customers.",'12.40', '3', 'uploads\egg.png',2),
    ('12 Inch Veggie Patty',"If you're a vegetarian or even a carnivore with a taste for green, you have to try this tasty vegetarian selection. The patty is made from vegetables and brown rice and its own unique spices.It's one of our favorites.", '14.60', '3', 'uploads\veggie.png',2),
    ('Italian B.M.T Salad',"Salad", '12.40','3',"uploads\italian.jpg",2),
    ('Bulgogi Chicken Salad',"Salad", '10.60','3',"uploads\Bulgogi.jpg",2),
    ('Egg Mayo Salad',"Salad", '9.10','3',"uploads\egg.jpg",2),
    ('Veggie Patty Salad',"Salad", '11.30','3',"uploads\veggie.jpg",2),
    ("Chocolate Croissant", "Croissant", '1.80','1',"uploads\chocolate.jpg",3),("Apple Lattice", "Lattice", '1.80','1',"uploads\apple.jpg",3),("Peach Tart", "Tart", '2.50','1',"uploads\peach.jpg",3),
    ("Strawberry Tart", "Tart", '2.50','1',"uploads\strawberry.jpg",3),("Chicken Pie", "Pie", '1.80','1',"uploads\chicken.jpg",3),("Curry Puff", "Puff", '1.80','1',"uploads\curry.jpg",3),
    ("Tuna Puff", "Puff", '1.80','1',"uploads\tuna.jpg",3),("Vegetable Pie", "Pie", '1.80','1',"uploads\vegetable.jpg",3);

-- Create the addons table
CREATE TABLE addons (
    id INT AUTO_INCREMENT PRIMARY KEY,
    food_id INT,
    name VARCHAR(255),
    price VARCHAR(50),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (food_id) REFERENCES food_items(id) ON DELETE CASCADE
);
INSERT INTO addons (name, price, food_id)
VALUES 
    -- item1
    ('Lettuce', '0.0', 1), ('Tomatoes', '0.0', 1), ('Cucumbers', '0.0', 1), ('Green Capsicum', '0.0', 1),
    ('Onions', '0.0', 1), ('Olives', '0.0', 1), ('Pickles', '0.0', 1), ('Jalapenos', '0.0', 1),
    ('Sweet Corn', '0.0', 1), ('Spicy Mayo', '0.0', 1), ('Chilli', '0.0', 1), ('Mayonnaise', '0.0', 1),
    ('Honey Mustard', '0.0', 1), ('2 Extra Cheddar Slice', '0.80', 1), ('Extra Mozzarella', '0.80', 1),
    ('Extra Tuna', '1.70', 1), ('Extra Avocado', '1.70', 1), ('Extra Chopped Mushroom', '1.70', 1),
    ('Extra Chicken Bacon', '1.70', 1), ('Extra Egg Mayo', '1.70', 1),
    -- item2 
    ('Lettuce', '0.0', 2), ('Tomatoes', '0.0', 2), ('Cucumbers', '0.0', 2), ('Green Capsicum', '0.0', 2),
    ('Onions', '0.0', 2), ('Olives', '0.0', 2), ('Pickles', '0.0', 2), ('Jalapenos', '0.0', 2),
    ('Sweet Corn', '0.0', 2), ('Spicy Mayo', '0.0', 2), ('Chilli', '0.0', 2), ('Mayonnaise', '0.0', 2),
    ('Honey Mustard', '0.0', 2), ('2 Extra Cheddar Slice', '0.80', 2), ('Extra Mozzarella', '0.80', 2),
    ('Extra Tuna', '1.70', 2), ('Extra Avocado', '1.70', 2), ('Extra Chopped Mushroom', '1.70', 2),
    ('Extra Chicken Bacon', '1.70', 2), ('Extra Egg Mayo', '1.70', 2),
    -- item3
    ('Lettuce', '0.0', 3), ('Tomatoes', '0.0', 3), ('Cucumbers', '0.0', 3), ('Green Capsicum', '0.0', 3),
    ('Onions', '0.0', 3), ('Olives', '0.0', 3), ('Pickles', '0.0', 3), ('Jalapenos', '0.0', 3),
    ('Sweet Corn', '0.0', 3), ('Spicy Mayo', '0.0', 3), ('Chilli', '0.0', 3), ('Mayonnaise', '0.0', 3),
    ('Honey Mustard', '0.0', 3), ('2 Extra Cheddar Slice', '0.80', 3), ('Extra Mozzarella', '0.80', 3),
    ('Extra Tuna', '1.70', 3), ('Extra Avocado', '1.70', 3), ('Extra Chopped Mushroom', '1.70', 3),
    ('Extra Chicken Bacon', '1.70', 3), ('Extra Egg Mayo', '1.70', 3),
    -- item4
    ('Lettuce', '0.0', 4), ('Tomatoes', '0.0', 4), ('Cucumbers', '0.0', 4), ('Green Capsicum', '0.0', 4),
    ('Onions', '0.0', 4), ('Olives', '0.0', 4), ('Pickles', '0.0', 4), ('Jalapenos', '0.0', 4),
    ('Sweet Corn', '0.0', 4), ('Spicy Mayo', '0.0', 4), ('Chilli', '0.0', 4), ('Mayonnaise', '0.0', 4),
    ('Honey Mustard', '0.0', 4), ('2 Extra Cheddar Slice', '0.80', 4), ('Extra Mozzarella', '0.80', 4),
    ('Extra Tuna', '1.70', 4), ('Extra Avocado', '1.70', 4), ('Extra Chopped Mushroom', '1.70', 4),
    ('Extra Chicken Bacon', '1.70', 4), ('Extra Egg Mayo', '1.70', 4),
    -- item5
    ('Lettuce', '0.0', 5), ('Tomatoes', '0.0', 5), ('Cucumbers', '0.0', 5), ('Green Capsicum', '0.0', 5),
    ('Onions', '0.0', 5), ('Olives', '0.0', 5), ('Pickles', '0.0', 5), ('Jalapenos', '0.0', 5),
    ('Sweet Corn', '0.0', 5), ('Spicy Mayo', '0.0', 5), ('Chilli', '0.0', 5), ('Mayonnaise', '0.0', 5),
    ('Honey Mustard', '0.0', 5), ('2 Extra Cheddar Slice', '0.80', 5), ('Extra Mozzarella', '0.80', 5),
    ('Extra Tuna', '1.70', 5), ('Extra Avocado', '1.70', 5), ('Extra Chopped Mushroom', '1.70', 5),
    ('Extra Chicken Bacon', '1.70', 5), ('Extra Egg Mayo', '1.70', 5),
    -- item6
    ('Lettuce', '0.0', 6), ('Tomatoes', '0.0', 6), ('Cucumbers', '0.0', 6), ('Green Capsicum', '0.0', 6),
    ('Onions', '0.0', 6), ('Olives', '0.0', 6), ('Pickles', '0.0', 6), ('Jalapenos', '0.0', 6),
    ('Sweet Corn', '0.0', 6), ('Spicy Mayo', '0.0', 6), ('Chilli', '0.0', 6), ('Mayonnaise', '0.0', 6),
    ('Honey Mustard', '0.0', 6), ('2 Extra Cheddar Slice', '0.80', 6), ('Extra Mozzarella', '0.80', 6),
    ('Extra Tuna', '1.70', 6), ('Extra Avocado', '1.70', 6), ('Extra Chopped Mushroom', '1.70', 6),
    ('Extra Chicken Bacon', '1.70', 6), ('Extra Egg Mayo', '1.70', 6),
    -- item7
    ('Lettuce', '0.0', 7), ('Tomatoes', '0.0', 7), ('Cucumbers', '0.0', 7), ('Green Capsicum', '0.0', 7),
    ('Onions', '0.0', 7), ('Olives', '0.0', 7), ('Pickles', '0.0', 7), ('Jalapenos', '0.0', 7),
    ('Sweet Corn', '0.0', 7), ('Spicy Mayo', '0.0', 7), ('Chilli', '0.0', 7), ('Mayonnaise', '0.0', 7),
    ('Honey Mustard', '0.0', 7), ('2 Extra Cheddar Slice', '0.80', 7), ('Extra Mozzarella', '0.80', 7),
    ('Extra Tuna', '1.70', 7), ('Extra Avocado', '1.70', 7), ('Extra Chopped Mushroom', '1.70', 7),
    ('Extra Chicken Bacon', '1.70', 7), ('Extra Egg Mayo', '1.70', 7),
    -- item8
    ('Lettuce', '0.0', 8), ('Tomatoes', '0.0', 8), ('Cucumbers', '0.0', 8), ('Green Capsicum', '0.0', 8),
    ('Onions', '0.0', 8), ('Olives', '0.0', 8), ('Pickles', '0.0', 8), ('Jalapenos', '0.0', 8),
    ('Sweet Corn', '0.0', 8), ('Spicy Mayo', '0.0', 8), ('Chilli', '0.0', 8), ('Mayonnaise', '0.0', 8),
    ('Honey Mustard', '0.0', 8), ('2 Extra Cheddar Slice', '0.80', 8), ('Extra Mozzarella', '0.80', 8),
    ('Extra Tuna', '1.70', 8), ('Extra Avocado', '1.70', 8), ('Extra Chopped Mushroom', '1.70', 8),
    ('Extra Chicken Bacon', '1.70', 8), ('Extra Egg Mayo', '1.70', 8),
    -- item9
    ('Lettuce', '0.0', 9), ('Tomatoes', '0.0', 9), ('Cucumbers', '0.0', 9), ('Green Capsicum', '0.0', 9),
    ('Onions', '0.0', 9), ('Olives', '0.0', 9), ('Pickles', '0.0', 9), ('Jalapenos', '0.0', 9),
    ('Sweet Corn', '0.0', 9), ('Spicy Mayo', '0.0', 9), ('Chilli', '0.0', 9), ('Mayonnaise', '0.0', 9),
    ('Honey Mustard', '0.0', 9), ('2 Extra Cheddar Slice', '0.80', 9), ('Extra Mozzarella', '0.80', 9),
    ('Extra Tuna', '1.70', 9), ('Extra Avocado', '1.70', 9), ('Extra Chopped Mushroom', '1.70', 9),
    ('Extra Chicken Bacon', '1.70', 9), ('Extra Egg Mayo', '1.70', 9),

    -- item10
    ('Lettuce', '0.0', 10), ('Tomatoes', '0.0', 10), ('Cucumbers', '0.0', 10), ('Green Capsicum', '0.0', 10),
    ('Onions', '0.0', 10), ('Olives', '0.0', 10), ('Pickles', '0.0', 10), ('Jalapenos', '0.0', 10),
    ('Sweet Corn', '0.0', 10), ('Spicy Mayo', '0.0', 10), ('Chilli', '0.0', 10), ('Mayonnaise', '0.0', 10),
    ('Honey Mustard', '0.0', 10), ('2 Extra Cheddar Slice', '0.80', 10), ('Extra Mozzarella', '0.80', 10),
    ('Extra Tuna', '1.70', 10), ('Extra Avocado', '1.70', 10), ('Extra Chopped Mushroom', '1.70', 10),
    ('Extra Chicken Bacon', '1.70', 10), ('Extra Egg Mayo', '1.70', 10),

    -- item11
    ('Lettuce', '0.0', 11), ('Tomatoes', '0.0', 11), ('Cucumbers', '0.0', 11), ('Green Capsicum', '0.0', 11),
    ('Onions', '0.0', 11), ('Olives', '0.0', 11), ('Pickles', '0.0', 11), ('Jalapenos', '0.0', 11),
    ('Sweet Corn', '0.0', 11), ('Spicy Mayo', '0.0', 11), ('Chilli', '0.0', 11), ('Mayonnaise', '0.0', 11),
    ('Honey Mustard', '0.0', 11), ('2 Extra Cheddar Slice', '0.80', 11), ('Extra Mozzarella', '0.80', 11),
    ('Extra Tuna', '1.70', 11), ('Extra Avocado', '1.70', 11), ('Extra Chopped Mushroom', '1.70', 11),
    ('Extra Chicken Bacon', '1.70', 11), ('Extra Egg Mayo', '1.70', 11),

    -- item12
    ('Lettuce', '0.0', 12), ('Tomatoes', '0.0', 12), ('Cucumbers', '0.0', 12), ('Green Capsicum', '0.0', 12),
    ('Onions', '0.0', 12), ('Olives', '0.0', 12), ('Pickles', '0.0', 12), ('Jalapenos', '0.0', 12),
    ('Sweet Corn', '0.0', 12), ('Spicy Mayo', '0.0', 12), ('Chilli', '0.0', 12), ('Mayonnaise', '0.0', 12),
    ('Honey Mustard', '0.0', 12), ('2 Extra Cheddar Slice', '0.80', 12), ('Extra Mozzarella', '0.80', 12),
    ('Extra Tuna', '1.70', 12), ('Extra Avocado', '1.70', 12), ('Extra Chopped Mushroom', '1.70', 12),
    ('Extra Chicken Bacon', '1.70', 12), ('Extra Egg Mayo', '1.70', 12);



-- Create Order table
CREATE TABLE orders (
    id INT AUTO_INCREMENT PRIMARY KEY,
    user_id INT NOT NULL,
    vendor_id INT NOT NULL,
    total_price DECIMAL(10,2) NOT NULL,
    pickup_time TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    status VARCHAR(50) NOT NULL DEFAULT 'Pending',
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
	FOREIGN KEY (vendor_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE TABLE order_items (
    id INT AUTO_INCREMENT PRIMARY KEY,
    order_id INT NOT NULL,
    food_id INT NOT NULL,
    quantity INT NOT NULL,
    addons_ids JSON,  
    item_price DECIMAL(10,2) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (order_id) REFERENCES orders(id) ON DELETE CASCADE,
    FOREIGN KEY (food_id) REFERENCES food_items(id) ON DELETE CASCADE
);