
interface SideNavLogoProps {
  href: string;
  img: string;
  imgAlt: string;
}

export const SideNavLogo = ({ href, img, imgAlt }: SideNavLogoProps) => {
  return (
    <a href={href} className="flex items-center p-4">
      <img src={img} alt={imgAlt} className="h-8" />
    </a>
  );
};
