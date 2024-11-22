// components/UpcomingGames.tsx
"use client";
import React, { useState, useEffect } from "react";
import GameInfo from "./gameInfo";

interface Game {
  opponent: string;
  dateTime: string;
  venue: string;
  lowestTicketPrice: string;
  homeTeam: string;
  awayTeam: string;
  winOdds?: number; // Optional, as it's not in the API response
  injuredPlayers?: string[]; // Optional, as it's not in the API response
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

  // Map of team abbreviations to full names
  const teamAbbreviations: Record<string, string> = {
    ATL: "Atlanta Hawks",
    BOS: "Boston Celtics",
    BKN: "Brooklyn Nets",
    CHA: "Charlotte Hornets",
    CHI: "Chicago Bulls",
    CLE: "Cleveland Cavaliers",
    DAL: "Dallas Mavericks",
    DEN: "Denver Nuggets",
    DET: "Detroit Pistons",
    GSW: "Golden State Warriors",
    HOU: "Houston Rockets",
    IND: "Indiana Pacers",
    LAC: "LA Clippers",
    LAL: "Los Angeles Lakers",
    MEM: "Memphis Grizzlies",
    MIA: "Miami Heat",
    MIL: "Milwaukee Bucks",
    MIN: "Minnesota Timberwolves",
    NOP: "New Orleans Pelicans",
    NYK: "New York Knicks",
    OKC: "Oklahoma City Thunder",
    ORL: "Orlando Magic",
    PHI: "Philadelphia 76ers",
    PHX: "Phoenix Suns",
    POR: "Portland Trail Blazers",
    SAC: "Sacramento Kings",
    SAS: "San Antonio Spurs",
    TOR: "Toronto Raptors",
    UTA: "Utah Jazz",
    WAS: "Washington Wizards",
  };

  const [games, setGames] = useState<Game[]>([]);

  // Function to fetch games data
  const fetchGames = async () => {
    try {
      const response = await fetch("http://localhost:8080/get", {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify({ Team: team }),
      });

      if (!response.ok) {
        throw new Error("Failed to fetch games");
      }

      const data = await response.json();
      // data.games is an array of game objects

      // Map API data to our Game interface
      const mappedGames: Game[] = data.games.map((gameData: any) => {
        const homeTeamAbbr = gameData.home_team;
        const awayTeamAbbr = gameData.away_team;

        const homeTeamFullName =
          teamAbbreviations[homeTeamAbbr] || homeTeamAbbr;
        const awayTeamFullName =
          teamAbbreviations[awayTeamAbbr] || awayTeamAbbr;

        // Determine opponent based on whether the team is home or away
        const opponent =
          team === homeTeamAbbr ? awayTeamFullName : homeTeamFullName;

        return {
          opponent: opponent,
          dateTime: gameData.start_time,
          venue: gameData.venueName,
          lowestTicketPrice: gameData.lowest_ticket_price,
          homeTeam: homeTeamFullName,
          awayTeam: awayTeamFullName,
          // winOdds and injuredPlayers can be added if available
        };
      });

      setGames(mappedGames);
    } catch (error) {
      console.error("Error fetching games:", error);
    }
  };

  useEffect(() => {
    // Fetch games when component mounts
    fetchGames();

    // Set up interval to fetch games every minute
    const interval = setInterval(() => {
      fetchGames();
    }, 60000); // 60000 ms = 1 minute

    // Cleanup interval on component unmount
    return () => clearInterval(interval);
  }, [team]);

  return (
    <div className="flex flex-col justify-center">
      {games.map((game) => {
        const opponentLogo = teamToLogo[game.opponent] || "default-logo.svg";

        return (
          <GameInfo
            key={`${game.opponent}-${game.dateTime}`}
            team={teamAbbreviations[team] || team}
            opponent={game.opponent}
            dateTime={game.dateTime}
            venue={game.venue}
            opponentLogo={opponentLogo}
            lowestTicketPrice={game.lowestTicketPrice}
            // winOdds and injuredPlayers can be passed if available
          />
        );
      })}
    </div>
  );
}

