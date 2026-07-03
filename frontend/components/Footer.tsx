import Link from "next/link";

export default function Footer() {
  return (
    <footer className="bg-ink text-[#94897b] px-6 md:px-16 pt-16 pb-8">
      <div className="grid grid-cols-2 md:grid-cols-[2fr_1fr_1fr_1fr] gap-8 md:gap-11 pb-10 border-b border-white/10 mb-7">
        <div className="col-span-2 md:col-span-1">
          <Link
            href="/"
            className="font-display text-xl tracking-[0.25em] uppercase text-warmwhite block mb-4"
          >
            SBY <span className="text-champ italic">TWILIGHT</span>
          </Link>
          <p className="text-xs leading-relaxed max-w-[270px] mb-4">
            Handcrafted Indian fashion — where every piece tells a story.
            Made to order, shipped across the world.
          </p>
          <div className="text-[11px] leading-loose">
            <a href="mailto:sby.twilight4@gmail.com" className="hover:text-clay transition-colors">
              sby.twilight4@gmail.com
            </a>
            <br />
            <a href="tel:+918591805622" className="hover:text-clay transition-colors">
              +91 85918 05622
            </a>
            <br />
            <span className="text-[10px] text-[#6b6259] leading-relaxed block mt-1">
              Vijay Estate, Kallu Pawala Chawal,<br />
              Halav Pool, Near Dhobi Ghat,<br />
              Kurla West, Mumbai 400070
            </span>
          </div>
        </div>
        <FooterCol
          heading="Shop"
          links={[
            ["Suit Sets", "/shop?category=suit-sets"],
            ["Anarkali", "/shop?category=anarkali"],
            ["Gharara", "/shop?category=gharara"],
            ["Co-ord Sets", "/shop?category=co-ord-sets"],
          ]}
        />
        <FooterCol
          heading="Account"
          links={[
            ["Sign In", "/account"],
            ["My Orders", "/account/orders"],
            ["Cart", "/cart"],
          ]}
        />
        <FooterCol
          heading="Support"
          links={[
            ["Shipping", "#"],
            ["Returns", "#"],
            ["Size Guide", "#"],
            ["FAQ", "#"],
          ]}
        />
      </div>
      <div className="flex flex-col md:flex-row justify-between items-center gap-3 text-[11px]">
        <span>© 2026 SBY TWILIGHT. All rights reserved.</span>
        <div className="flex gap-4">
          <a
            href="https://instagram.com/sby_twilight"
            target="_blank"
            rel="noopener noreferrer"
            className="hover:text-clay"
          >
            Instagram
          </a>
          <a href="#" className="hover:text-clay">
            Facebook
          </a>
          <a href="#" className="hover:text-clay">
            Pinterest
          </a>
        </div>
      </div>
    </footer>
  );
}

function FooterCol({
  heading,
  links,
}: {
  heading: string;
  links: [string, string][];
}) {
  return (
    <div>
      <h4 className="text-[10px] tracking-[0.22em] uppercase text-warmwhite mb-5">
        {heading}
      </h4>
      <ul className="space-y-2.5">
        {links.map(([label, href]) => (
          <li key={label}>
            <Link href={href} className="text-xs hover:text-clay transition-colors">
              {label}
            </Link>
          </li>
        ))}
      </ul>
    </div>
  );
}
