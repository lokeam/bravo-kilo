import { useAuth } from "../AuthContext";
import { User } from "../AuthContext";

// const isUser = (user: User | null): user is User => {
//   return (
//     user !== null &&
//     typeof user === 'object' &&
//     'firstName' in user &&
//     'last_name' in user &&
//     'picture' in user
//   );
// }

export default function Avatar() {
  console.log('Avatar root');
  const { user } = useAuth();

  console.log('User from autAuth: ',user);

  // if (!isUser(user)) return null;
  const { picture, firstName, lastName } = user;

  const createInitials = (firstName = 'N', lastName='A') => {
    return firstName[0]+lastName[0];
  }

  const userInitials = createInitials(firstName, lastName);

  // Development: Save sizes for Design audit
  const avatarSize = {
    'sm': 'h-10 w-10',
    'md': 'h-20 w-20',
    'lg': 'h-32 w-32'
  };

  console.log('Avatar component firstName: ', firstName);
  console.log('Avatar component userInitials: ', userInitials)




/*
        {
          picture === '' || picture === undefined ? (
            <div className="relative inline-flex items-center justify-center overflow-hidden bg-gray-600 h-10 w-10 rounded-full">
              <span className="font-medium text-gray-600 dark:text-gray-300">{userInitials}</span>
            </div>
          ) : (
            <img
              alt={`User avatar for ${firstName}`}
              className={`rounded-full ${avatarSize['sm']}`}
              src={picture}
            />
          )
        }

*/



  return (
    <div className="flex flex-row justify-center space-x-5 rounded avatar text-white">
      <div aria-label="Bravo Kilo user avatar" className="relative">
        <div className="relative inline-flex items-center justify-center overflow-hidden bg-gray-600 h-10 w-10 rounded-full">
          <span className="font-medium text-gray-600 dark:text-gray-300">{userInitials}</span>
        </div>
      </div>
      <div className="flex flex-col text-left">
          <div className="text-base font-bold">{user?.firstName}'s Login</div>
          <div className="text-nevada-gray text-sm">{user?.email}</div>
        </div>
    </div>
  )
}
