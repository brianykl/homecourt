import Image from "next/image";

export default function GameInfo( { team, logo }: { team: string, logo: string}) {
    return(
        <Image src={logo} height={100} width={100} alt="next game"/>
    )
}