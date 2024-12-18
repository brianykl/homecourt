"use client"
import BrianProfile from "@/components/brianProfile";
import Header from "@/components/header";
import TeamSelector from "@/components/teamSelector";
import UpcomingGames from "@/components/upcomingGames";
import { useState } from "react";


export default function Home() {
  const [team, setTeam] = useState("")
  return (
    <div className="flex flex-col w-screen">
      <div className="flex flex-row justify-between items-center" id="header">
        <BrianProfile/>
        <Header/>
        <TeamSelector selectedTeam={ team } setSelectedTeam={ setTeam }/>
      </div>
      <div className="flex flex-col justify-center" id="body">
        {team && <UpcomingGames team={team} />}
      </div>
      <div id="footer">
      </div>
    </div>
  );
}
