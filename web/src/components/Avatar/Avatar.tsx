
export interface AvatarProps {
  alt?: string;
  img?: string;
  initials: string;
  size?: 'sm' | 'md' | 'lg';
}

export default function Avatar({ alt, img, initials, size = 'sm'}: AvatarProps) {
  const truncatedInitials = initials.length > 2 ? initials.substring(0, 2) : initials;

  const avatarSize = {
    'sm': 'h-10 w-10',
    'md': 'h-20 w-20',
    'lg': 'h-32 w-32'
  };

  return (
    <div className="flex items-center justify-center space-x-4 rounded avatar text-white">
      <div aria-label="Kilo Bravo user avatar" className="relative">
        {
          img === '' ? (
            <div className="relative inline-flex items-center justify-center overflow-hidden bg-gray-600 h-10 w-10 rounded-full">
              <span className="font-medium text-gray-600 dark:text-gray-300">{truncatedInitials}</span>
            </div>
          ) : (
            <img
              alt={(alt === '') ? 'user avatar' : alt}
              className={`rounded-full ${avatarSize[size]}`}
              src={img}
            />
          )
        }
      </div>
    </div>
  )
}
