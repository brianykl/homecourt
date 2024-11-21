// components/GameInfo.tsx
"use client";
import React from "react";
import Image from "next/image";

interface GameInfoProps {
  team: string;
  opponent: string;
  winOdds: number;
  dateTime: string;
  venue: string;
  injuredPlayers: string[];
  opponentLogo: string;
}

const GameInfo: React.FC<GameInfoProps> = ({
  team,
  opponent,
  winOdds,
  dateTime,
  venue,
  injuredPlayers,
  opponentLogo,
}) => {
  // Format the date and time for better readability
  const formattedDate = new Date(dateTime).toLocaleDateString("en-US", {
    year: "numeric",
    month: "long",
    day: "numeric",
  });

  const formattedTime = new Date(dateTime).toLocaleTimeString("en-US", {
    hour: "2-digit",
    minute: "2-digit",
  });


  return (
    <div className="flex items-center p-4 border rounded-lg shadow-md bg-white">
      {/* Opponent Logo */}
      <div className="flex-shrink-0">
        <Image
          src={opponentLogo}
          alt={`${opponent} logo`}
          width={50}
          height={50}
        />
      </div>

      {/* event details */}
      <div className="ml-4 flex-1">
        <h2 className="text-xl font-bold">
          {team} vs {opponent}
        </h2>
        <p className="text-gray-600">
          <strong>Date:</strong> {formattedDate} at {formattedTime}
        </p>
        <p className="text-gray-600">
          <strong>Venue:</strong> {venue}
        </p>
      </div>
      {/* odds */}
      <div className="flex-1">
        <h3 className="text-xl font-bold">
          $30
        </h3>
      </div>
      {/* odds */}
      <div className="flex-1">
        <h3 className="text-xl font-bold">
          ~{winOdds}% to win
        </h3>
      </div>
      {/* injuries*/}
      <div className=" flex-1">
        <h3 className="text-xl font-bold">
          Kevin Durant
        </h3>
      </div>
    </div>
  );
};

export default GameInfo;