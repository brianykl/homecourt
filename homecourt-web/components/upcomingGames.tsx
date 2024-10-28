import GameInfo from "./gameInfo";

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
    }
    
    return(
        <div className="flex flex-col justify-center">
            <GameInfo team={team} logo={teamToLogo[team]} />
            <GameInfo team={team} logo={teamToLogo[team]} />
            <GameInfo team={team} logo={teamToLogo[team]} />
            <GameInfo team={team} logo={teamToLogo[team]} />
            <GameInfo team={team} logo={teamToLogo[team]} />
        </div>
    )
}