// components/GameInfo.tsx
"use client";
import React from "react";
import Image from "next/image";

interface GameInfoProps {
  team: string;
  opponent: string;
  dateTime: string;
  venue: string;
  opponentLogo: string;
  lowestTicketPrice: string;
  winOdds?: number;
  injuredPlayers?: string[];
}

const GameInfo: React.FC<GameInfoProps> = ({
  team,
  opponent,
  dateTime,
  venue,
  opponentLogo,
  lowestTicketPrice,
  winOdds,
  injuredPlayers,
}) => {
  // Format the date and time for better readability
  const date = new Date(dateTime);
  const formattedDate = date.toLocaleDateString("en-US", {
    year: "numeric",
    month: "long",
    day: "numeric",
  });

  const formattedTime = date.toLocaleTimeString("en-US", {
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

      {/* Event Details */}
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
      {/* Lowest Ticket Price */}
      <div className="flex-1">
        <h3 className="text-xl font-bold">{lowestTicketPrice}</h3>
      </div>
      {/* Win Odds */}
      {winOdds !== undefined && (
        <div className="flex-1">
          <h3 className="text-xl font-bold">~{winOdds}% to win</h3>
        </div>
      )}
      {/* Injured Players */}
      {injuredPlayers && injuredPlayers.length > 0 && (
        <div className="flex-1">
          <h3 className="text-xl font-bold">
            {injuredPlayers.join(", ")}
          </h3>
        </div>
      )}
    </div>
  );
};

export default GameInfo;
