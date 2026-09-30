USE maz_suplementos;

INSERT INTO categories (name,description) VALUES
('Proteína','Fuentes de proteína para complementar la alimentación'),('Creatina','Productos a base de creatina'),('Pre-entreno','Fórmulas para antes del entrenamiento'),('Aminoácidos','Aminoácidos y mezclas relacionadas'),('Vitaminas','Vitaminas y minerales'),('Ganadores de masa','Productos de mayor aporte energético'),('Hidratación','Electrolitos y apoyo para hidratación'),('Salud y bienestar','Suplementos de bienestar general')
ON DUPLICATE KEY UPDATE description=VALUES(description);
INSERT INTO goals (name,description) VALUES
('Ganancia muscular','Apoyo nutricional para etapas de ganancia'),('Fuerza','Complementos para rutinas de fuerza'),('Recuperación','Apoyo nutricional posterior al entrenamiento'),('Energía','Productos orientados a energía'),('Resistencia','Complementos para sesiones prolongadas'),('Pérdida de grasa','Complementos para planes de control de peso'),('Salud general','Nutrientes para bienestar general'),('Rendimiento deportivo','Productos para acompañar la práctica deportiva')
ON DUPLICATE KEY UPDATE description=VALUES(description);
INSERT INTO ingredients (name,description) VALUES
('Creatina monohidratada','Creatina en presentación monohidratada'),('Whey concentrate','Concentrado de proteína de suero'),('Whey isolate','Aislado de proteína de suero'),('Cafeína','Ingrediente estimulante; revisar la etiqueta'),('BCAA','Aminoácidos de cadena ramificada'),('Glutamina','Aminoácido presente en proteínas'),('Electrolitos','Mezcla de minerales'),('Vitaminas','Mezcla de vitaminas'),('Minerales','Mezcla de minerales'),('Sin azúcar','Característica de formulación'),('Vegano','Formulado sin ingredientes de origen animal declarados')
ON DUPLICATE KEY UPDATE description=VALUES(description);

INSERT INTO supplements (name,brand,description,price,stock,presentation,flavor,weight,active)
SELECT 'Maz Creatine 300 g','Maz Demo','Creatina monohidratada en polvo. Producto demostrativo; consulta siempre la etiqueta.',449.00,18,'Bote','Sin sabor','300 g',TRUE
WHERE NOT EXISTS (SELECT 1 FROM supplements WHERE name='Maz Creatine 300 g' AND brand='Maz Demo');
INSERT INTO supplements (name,brand,description,price,stock,presentation,flavor,weight,active)
SELECT 'Maz Whey Vanilla 2 lb','Maz Demo','Mezcla de proteína de suero sabor vainilla para complementar la alimentación.',799.00,12,'Bolsa','Vainilla','2 lb',TRUE
WHERE NOT EXISTS (SELECT 1 FROM supplements WHERE name='Maz Whey Vanilla 2 lb' AND brand='Maz Demo');
INSERT INTO supplements (name,brand,description,price,stock,presentation,flavor,weight,active)
SELECT 'Maz Pre-Workout Citrus','Maz Demo','Pre-entreno sabor cítrico con cafeína. Revisa ingredientes y porciones antes de consumir.',529.00,5,'Bote','Cítricos','250 g',TRUE
WHERE NOT EXISTS (SELECT 1 FROM supplements WHERE name='Maz Pre-Workout Citrus' AND brand='Maz Demo');
INSERT INTO supplements (name,brand,description,price,stock,presentation,flavor,weight,active)
SELECT 'Maz Electrolytes','Maz Demo','Mezcla en polvo con electrolitos para preparar una bebida.',279.00,20,'Bolsa','Limón','20 porciones',TRUE
WHERE NOT EXISTS (SELECT 1 FROM supplements WHERE name='Maz Electrolytes' AND brand='Maz Demo');

INSERT IGNORE INTO supplement_categories SELECT s.id,c.id FROM supplements s JOIN categories c ON (s.name='Maz Creatine 300 g' AND c.name='Creatina') OR (s.name='Maz Whey Vanilla 2 lb' AND c.name='Proteína') OR (s.name='Maz Pre-Workout Citrus' AND c.name='Pre-entreno') OR (s.name='Maz Electrolytes' AND c.name='Hidratación') WHERE s.brand='Maz Demo';
INSERT IGNORE INTO supplement_goals SELECT s.id,g.id FROM supplements s JOIN goals g ON (s.name='Maz Creatine 300 g' AND g.name IN ('Fuerza','Rendimiento deportivo')) OR (s.name='Maz Whey Vanilla 2 lb' AND g.name IN ('Ganancia muscular','Recuperación')) OR (s.name='Maz Pre-Workout Citrus' AND g.name IN ('Energía','Rendimiento deportivo')) OR (s.name='Maz Electrolytes' AND g.name IN ('Resistencia','Recuperación')) WHERE s.brand='Maz Demo';
INSERT IGNORE INTO supplement_ingredients SELECT s.id,i.id FROM supplements s JOIN ingredients i ON (s.name='Maz Creatine 300 g' AND i.name='Creatina monohidratada') OR (s.name='Maz Whey Vanilla 2 lb' AND i.name IN ('Whey concentrate','Whey isolate')) OR (s.name='Maz Pre-Workout Citrus' AND i.name='Cafeína') OR (s.name='Maz Electrolytes' AND i.name IN ('Electrolitos','Sin azúcar')) WHERE s.brand='Maz Demo';
