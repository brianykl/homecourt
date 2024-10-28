import Image from "next/image";

export default function Header() {
    return(
        <div className="flex flex-row justify-self-start">
          <Image
            src="/homecourt.png"
            width={150}
            height={150}
            alt="homecourt logo"
          />
          <h1 className="self-center text-5xl font-bold"> homecourt </h1>
        </div>
    )
}