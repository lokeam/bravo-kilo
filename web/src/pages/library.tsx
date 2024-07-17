import { useAuth } from "../components/AuthContext";

const Library = () => {
  const { logout } = useAuth();

  return (
    <div className="bk_lib">
      <h1>Library</h1>

      <button onClick={logout}>Sign out of your Kilo Bravo account</button>
    </div>
  )
}

export default Library;
