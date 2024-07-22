import { Outlet, useLocation } from "react-router-dom";
import { useQuery } from "@tanstack/react-query";
import { useAuth } from "../components/AuthContext";
import axios from "axios";

import TopNavigation from "../components/TopNav/TopNav";
import SideNavigation from "../components/SideNav/SideNavigation";
import CardList from "../components/CardList/CardList";

interface Book {
  authors: string[];
  imageLinks: string[];
  title: string;
  subtitle: string;
  details: {
    genres: string[];
    description: string;
    isbn10: string;
    isbn13: string;
    language: string;
    pageCount: number;
    publishDate: number;
  };
}

const fetchUserBooks = async () => {
  const { data } = await axios.get(`${import.meta.env.VITE_API_ENDPOINT}/api/v1/user/books`, {
    withCredentials: true
  });
  return data.books;
};

const Library = () => {
  const { logout } = useAuth();
  const { search } = useLocation();
  const query = new URLSearchParams(search);
  const userID = parseInt(query.get('userID') || '0', 10);

  const { data: books, isLoading, isError } = useQuery({
    queryKey: ['userBooks'],
    queryFn: fetchUserBooks,
    enabled: !!userID,
  });

  if (isLoading) {
    return <div>Loading...</div>;
  }

  if (isError) {
    return <div>Error loading books</div>;
  }

  console.log('books:', books)

  return (
    <div className="bk_lib flex flex-col place-content-around lg:px-8 antialiased md:ml-64 h-screen pt-24">
      <TopNavigation />
      <SideNavigation />

      <h1>Library</h1>
      <button onClick={logout}>Sign out of your Kilo Bravo account</button>

      <CardList books={books} />

      <Outlet />
    </div>
  )
}

export default Library;
