import { Suspense } from 'react';
import { Route, Routes } from 'react-router-dom';
import { AppProvider } from './components/AuthContext';
import ProtectedRoute from './components/ProtectedRoute';
import Home from './pages/Home';
import EditBook from './pages/EditBook';
import AddBook from './pages/AddBook';
import ManualAdd from './pages/ManualAdd';
import SearchAdd from './pages/SearchAdd';
import Login from './pages/Login';
import Library from './pages/Library';
import AuthorGenre from './pages/AuthorGenre';
import BookDetail from './pages/BookDetail';

import NotFound from './pages/NotFound';

import './App.css'
import AuthenticatedLayout from './pages/AuthLayout';


function App() {
  return (
    <AppProvider>
      <Suspense fallback={<h1>Loading...</h1>}>
        <Routes>
          <Route path="/" element={<Home />} />
          <Route path="/login" element={<Login />}/>
          <Route element={<ProtectedRoute><AuthenticatedLayout /></ProtectedRoute>}>
            <Route path="/home" element={<Home />} />
            <Route path="/library" element={<Library />} />
            <Route path="/library/:authorID" element={<AuthorGenre />} /> {/* Dynamic route */}
            <Route path="/library/books/add" element={<AddBook />} />
            <Route path="/library/books/add/manual" element={<ManualAdd /> }/>
            <Route path="/library/books/add/search" element={<SearchAdd />} />
            <Route path="/library/books/:bookID" element={<BookDetail />} />
            <Route path="/library/books/:bookID/edit" element={<EditBook />} />
          </Route>
          <Route path="*" element={<NotFound />}/>
        </Routes>
      </Suspense>
    </AppProvider>
  );
}

export default App;
