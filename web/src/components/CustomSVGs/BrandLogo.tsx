
type BrandLogoProps = {
  className: string;
}

function BrandLogo({ className }: BrandLogoProps ) {
  return (
    <svg className={className} xmlns="http://www.w3.org/2000/svg"  viewBox="60 260 140 80">
      <defs>
        <linearGradient id="a" x1="137.09" x2="126.62" y1="346.22" y2="316.91" gradientUnits="userSpaceOnUse">
          <stop offset=".04" stopColor="#674099"/>
          <stop offset="1" stopColor="#482978"/>
        </linearGradient>
        <linearGradient id="b" x1="137.13" x2="121.43" y1="327.17" y2="280.06" gradientUnits="userSpaceOnUse">
          <stop offset=".04" stopColor="#3d6fb6"/>
          <stop offset="1" stopColor="#2b65b0"/>
        </linearGradient>
        <linearGradient id="c" x1="120.03" x2="171.85" y1="271.6" y2="331.28" gradientUnits="userSpaceOnUse">
          <stop offset=".04" stopColor="#85c558"/>
          <stop offset="1" stopColor="#58b947"/>
        </linearGradient>
      </defs>
      <path fill="url(#a)" d="m197.09 318.34-43.39-29.12a8.602 8.602 0 0 0-8.86-.44L67.4 330.19c-5.73 3.06-6.13 11.12-.74 14.74l30.75 20.64a8.602 8.602 0 0 0 8.86.44l30.37-16.24a8.585 8.585 0 0 1 5.26-.93l16.98 2.4c1.81.26 3.65-.07 5.26-.93l32.2-17.22c5.73-3.06 6.13-11.12.74-14.74Z"/>
      <path fill="url(#b)" d="m195.55 293.49-43.39-29.12a8.602 8.602 0 0 0-8.86-.44l-77.44 41.41c-5.73 3.06-6.13 11.12-.74 14.74l30.75 20.64a8.602 8.602 0 0 0 8.86.44l30.37-16.24a8.585 8.585 0 0 1 5.26-.93l16.98 2.4c1.81.26 3.65-.07 5.26-.93l32.2-17.22c5.73-3.06 6.13-11.12.74-14.74Z"/>
      <path fill="url(#c)" d="m194.23 268.65-43.39-29.12a8.602 8.602 0 0 0-8.86-.44L64.54 280.5c-5.73 3.06-6.13 11.12-.74 14.74l30.75 20.64a8.602 8.602 0 0 0 8.86.44l30.37-16.24a8.585 8.585 0 0 1 5.26-.93l16.98 2.4c1.81.26 3.65-.07 5.26-.93l32.2-17.22c5.73-3.06 6.13-11.12.74-14.74Z"/>
    </svg>
  )
}

export default BrandLogo;
