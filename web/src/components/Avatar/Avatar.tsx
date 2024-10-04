import { useAuth } from "../AuthContext";


export default function Avatar() {
  console.log('Avatar root');
  const { user } = useAuth();

  console.log('User from autAuth: ', user);

  const firstName = user?.firstName ?? 'N';
  const lastName = user?.lastName ?? 'A';

  const createInitials = (firstName = 'N', lastName='A') => {
    return firstName[0]+lastName[0];
  }

  const userInitials = createInitials(firstName, lastName);

  console.log('Avatar component firstName: ', firstName);
  console.log('Avatar component userInitials: ', userInitials)


  return (
    <div className="flex flex-row justify-center space-x-5 rounded avatar text-dark-ebony dark:text-white">
      <div aria-label="Bravo Kilo user avatar" className="relative">
        <div className="relative inline-flex items-center justify-center overflow-hidden bg-gray-300 dark:bg-gray-600 h-10 w-10 rounded-full">
          <span className="font-bold text-gray-600 dark:text-gray-300">{userInitials}</span>
        </div>
      </div>
      <div className="flex flex-col text-left">
          <div className="text-base font-bold">{user?.firstName}'s Login</div>
          <div className="text-dark-gunmetal dark:text-nevada-gray text-sm">{user?.email}</div>
        </div>
    </div>
  )
}
