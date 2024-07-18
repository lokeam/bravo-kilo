import { useAuth } from "../AuthContext";
import { User } from "../AuthContext";

const isUser = (user: User | null): user is User => {
  return (
    user !== null &&
    typeof user === 'object' &&
    'first_name' in user &&
    'last_name' in user &&
    'picture' in user
  );
}

export default function Avatar() {
  const { user } = useAuth();

  if (!isUser(user)) return null;

  const { picture, first_name, last_name } = user;

  const createInitials = (firstName = 'N', lastName='A') => {
    return firstName[0]+lastName[0];
  }

  const userInitials = createInitials(first_name, last_name);

  // Development: Save sizes for Design audit
  const avatarSize = {
    'sm': 'h-10 w-10',
    'md': 'h-20 w-20',
    'lg': 'h-32 w-32'
  };

  return (
    <div className="flex items-center justify-center space-x-4 rounded avatar text-white">
      <div aria-label="Bravo Kilo user avatar" className="relative">
        {
          picture === '' ? (
            <div className="relative inline-flex items-center justify-center overflow-hidden bg-gray-600 h-10 w-10 rounded-full">
              <span className="font-medium text-gray-600 dark:text-gray-300">{userInitials}</span>
            </div>
          ) : (
            <img
              alt={`User avatar for ${first_name}`}
              className={`rounded-full ${avatarSize['sm']}`}
              src={picture}
            />
          )
        }
      </div>
    </div>
  )
}
