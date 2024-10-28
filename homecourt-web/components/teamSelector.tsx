"use client";
import React, { FC } from "react"

interface TeamSelectorProps {
  selectedTeam: string
  setSelectedTeam: (team: string) => void
}

const TeamSelector: FC<TeamSelectorProps> = ({ selectedTeam, setSelectedTeam }) => {
  const teams = [
    "Atlanta Hawks",
    "Boston Celtics",
    "Brooklyn Nets",
    "Charlotte Hornets",
    "Chicago Bulls",
    "Cleveland Cavaliers",
    "Dallas Mavericks",
    "Denver Nuggets",
    "Detroit Pistons",
    "Golden State Warriors",
    "Houston Rockets",
    "Indiana Pacers",
    "LA Clippers",
    "Los Angeles Lakers",
    "Memphis Grizzlies",
    "Miami Heat",
    "Milwaukee Bucks",
    "Minnesota Timberwolves",
    "New Orleans Pelicans",
    "New York Knicks",
    "Oklahoma City Thunder",
    "Orlando Magic",
    "Philadelphia 76ers",
    "Phoenix Suns",
    "Portland Trail Blazers",
    "Sacramento Kings",
    "San Antonio Spurs",
    "Toronto Raptors",
    "Utah Jazz",
    "Washington Wizards",
  ];

  return (
    <div className="flex items-center pr-10 font-bold">
      <label
        htmlFor="team-select"
        className="mr-2 text-xl font-semibold"
      ></label>
      <select
        id="team-select"
        className="p-2 border-2 border-black rounded-xl"
        value={selectedTeam}
        onChange={(e) => setSelectedTeam(e.target.value)}
      >
        <option value="">select a team</option>
        {teams.map((team) => (
          <option key={team} value={team}>
            {team}
          </option>
        ))}
      </select>
    </div>
  )
}

export default TeamSelector;