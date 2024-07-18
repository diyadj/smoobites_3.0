drop schema smoobites;
create schema smoobites;
use smoobites;

CREATE TABLE users (
    id INT AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    email VARCHAR(100) UNIQUE NOT NULL,
    password VARCHAR(255) NOT NULL,
    role VARCHAR(50) NOT NULL
);
insert into users (name, email, password, role)
values ("khoon coffeehouse express", "khoon@smu.com", "$2a$10$g6C1dZ2Sa/MCye.pJW55J.OoIcg9bCwa.71jm7ZqT4WAJWuWfhD3S", "vendor");


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

-- Create the addons table
CREATE TABLE addons (
    id INT AUTO_INCREMENT PRIMARY KEY,
    food_id INT,
    name VARCHAR(255),
    price VARCHAR(50),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (food_id) REFERENCES food_items(id) ON DELETE CASCADE
);

--Create Order table
CREATE TABLE order(
    ID INT AUTO_INCREMENT PRIMARY KEY,
    user_id INT NOT NULL,
    vendor_id INT NOT NULL,
    food_id INT NOT NULL,
    addons_id INT,
    total_price DECIMAL(10,2) NOT NULL,
    pickup_time TIMESTAMP DEFAULT,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
    FOREIGN KEY (vendor_id) REFERENCES users(id) ON DELETE CASCADE
    FOREIGN KEY (food_id) REFERENCES food_items(id) ON DELETE CASCADE
    FOREIGN KEY (addons_id) REFERENCES addons(id) ON DELETE CASCADE

)
