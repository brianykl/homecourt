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
  teamLogo: string;
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
  teamLogo,
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
    <div className="flex items-center border p-2 rounded-lg shadow-md bg-white">
      {/* Opponent Logo */}
      <div className="flex-shrink-0 ml-2 mr-1">
        <Image
          src={teamLogo}
          alt={`${team} logo`}
          width={50}
          height={50}
        />
      </div>
      vs.
      <div className="flex-shrink-0 ml-0">
        <Image
          src={opponentLogo}
          alt={`${opponent} logo`}
          width={50}
          height={50}
        />
      </div>
      

      {/* Event Details */}
      <div className="ml-10 flex-1">
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
      <div className="w-32">
        <h3 className="text-xl font-bold">{lowestTicketPrice}</h3>
      </div>
      {/* Win Odds */}
      {winOdds !== undefined && (
        <div className="w-32">
          <h3 className="text-xl font-bold">~{winOdds}% to win</h3>
        </div>
      )}
      {winOdds == undefined && (
        <div className="w-32">
          <h3 className="text-xl font-bold">odds not avail.</h3>
        </div>
      )}
      {/* Injured Players */}
      {injuredPlayers && injuredPlayers.length > 0 && (
        <div className="w-32">
          <h3 className="text-xl font-bold">
            {injuredPlayers.join(", ")}
          </h3>
        </div>
      )}
      {(!injuredPlayers || (injuredPlayers && injuredPlayers.length == 0)) && (
        <div className="w-32">
          <h3 className="text-xl font-bold">
            no injuries
          </h3>
        </div>
      )}
    </div>
  );
};

export default GameInfo;
