Teams
  ├── team_id (PK)
  ├── name
  ├── city
  ├── league
  └── logo_url

Players
  ├── player_id (PK)
  ├── team_id (FK to Teams)
  ├── name
  └── status (e.g., Active, Injured)

Games
  ├── game_id (PK)
  ├── home_team_id (FK to Teams)
  ├── away_team_id (FK to Teams)
  ├── scheduled_date
  └── venue

Odds
  ├── odds_id (PK)
  ├── game_id (FK to Games)
  ├── home_team_odds
  ├── away_team_odds
  └── updated_at
