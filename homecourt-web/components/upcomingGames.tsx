import GameInfo from "./gameInfo";

interface Game {
  opponent: string; // The team you are playing against
  winOdds: number; // Probability of your team winning (e.g., 75 for 75%)
  dateTime: string; // Date and time of the game (ISO string or formatted date)
  venue: string; // Location of the game
  injuredPlayers: string[]; // List of injured players from both teams
}

export default function UpcomingGames({ team }: { team: string }) {
  const teamToLogo: Record<string, string> = {
    "Atlanta Hawks": "hawks.svg",
    "Boston Celtics": "celtics.svg",
    "Brooklyn Nets": "nets.svg",
    "Charlotte Hornets": "hornets.svg",
    "Chicago Bulls": "bulls.svg",
    "Cleveland Cavaliers": "cavs.svg",
    "Dallas Mavericks": "mavs.svg",
    "Denver Nuggets": "nuggets.svg",
    "Detroit Pistons": "pistons.svg",
    "Golden State Warriors": "warriors.svg",
    "Houston Rockets": "rockets.svg",
    "Indiana Pacers": "pacers.svg",
    "LA Clippers": "clippers.svg",
    "Los Angeles Lakers": "lakers.svg",
    "Memphis Grizzlies": "grizzlies.svg",
    "Miami Heat": "heat.svg",
    "Milwaukee Bucks": "bucks.svg",
    "Minnesota Timberwolves": "wolves.svg",
    "New Orleans Pelicans": "pelicans.svg",
    "New York Knicks": "knicks.svg",
    "Oklahoma City Thunder": "thunder.svg",
    "Orlando Magic": "magic.svg",
    "Philadelphia 76ers": "sixers.svg",
    "Phoenix Suns": "suns.svg",
    "Portland Trail Blazers": "blazers.svg",
    "Sacramento Kings": "kings.svg",
    "San Antonio Spurs": "spurs.svg",
    "Toronto Raptors": "raptors.svg",
    "Utah Jazz": "jazz.svg",
    "Washington Wizards": "wizards.svg",
  };

  // want to make our api call to backend redis to get latest game info
  // should be able go give team, and then get back information about upcoming five home games
//   const upcomingGameInfo = await fetch("http://localhost:8080/some_api");
  const mockUpcomingGames: Game[] = [
    {
      opponent: "Los Angeles Lakers",
      winOdds: 65,
      dateTime: "2024-11-05T19:30:00",
      venue: "Staples Center",
      injuredPlayers: [
        "LeBron James (Lakers)",
        "Anthony Davis (Lakers)",
        "Luka Doncic (Mavericks)",
      ],
    },
    {
      opponent: "Boston Celtics",
      winOdds: 70,
      dateTime: "2024-11-07T20:00:00",
      venue: "Staples Center",
      injuredPlayers: [
        "Jayson Tatum (Celtics)",
        "Jaylen Brown (Celtics)",
        "Kyrie Irving (Nets)",
      ],
    },
    {
      opponent: "Chicago Bulls",
      winOdds: 80,
      dateTime: "2024-11-10T18:00:00",
      venue: "Staples Center",
      injuredPlayers: [
        "Zach LaVine (Bulls)",
        "Nikola Vucevic (Bulls)",
        "Jimmy Butler (Heat)",
      ],
    },
    {
      opponent: "Miami Heat",
      winOdds: 60,
      dateTime: "2024-11-12T19:00:00",
      venue: "Staples Center",
      injuredPlayers: [
        "Bam Adebayo (Heat)",
        "Tyler Herro (Heat)",
        "Kawhi Leonard (Clippers)",
      ],
    },
    {
      opponent: "Golden State Warriors",
      winOdds: 55,
      dateTime: "2024-11-15T21:00:00",
      venue: "Staples Center",
      injuredPlayers: [
        "Stephen Curry (Warriors)",
        "Klay Thompson (Warriors)",
        "Draymond Green (Warriors)",
      ],
    },
  ];

  return (
    <div className="flex flex-col justify-center">
      {mockUpcomingGames.map((game) => {
        const opponentLogo = teamToLogo[game.opponent] || "default-logo.svg"; // Provide a default logo if not found

        return (
          <GameInfo
            key={`${game.opponent}-${game.dateTime}`} // Use a unique key
            team={team}
            opponent={game.opponent}
            winOdds={game.winOdds}
            dateTime={game.dateTime}
            venue={game.venue}
            injuredPlayers={game.injuredPlayers}
            opponentLogo={opponentLogo}
          />
        );
      })}
    </div>
  );
}
