-- GET /backup/supplement_goals; pagination enabled, max rows 2000
USE maz_suplementos; SELECT * FROM supplement_goals ORDER BY supplement_id,goal_id;
