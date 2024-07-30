import { Suspense } from 'react';
import { Route, Routes } from 'react-router-dom';
import { AuthProvider } from './components/AuthContext';
import ProtectedRoute from './components/ProtectedRoute';
import Home from './pages/Home';
import EditBook from './pages/EditBook';
import AddBook from './pages/AddBook';
import Login from './pages/Login';
import Library from './pages/Library';
import BookDetail from './pages/BookDetail';

import NotFound from './pages/NotFound';

import './App.css'
import AuthenticatedLayout from './pages/AuthLayout';


function App() {
  return (
    <AuthProvider>
      <Suspense fallback={<h1>Loading...</h1>}>
        <Routes>
          <Route path="/" element={<Home />} />
          <Route path="/login" element={<Login />}/>
          <Route element={<ProtectedRoute><AuthenticatedLayout /></ProtectedRoute>}>
            <Route path="/home" element={<Home />} />
            <Route path="/library" element={<Library />} />
            <Route path="/library/books/add" element={<AddBook />} />
            <Route path="/library/books/:bookID" element={<BookDetail />} />
            <Route path="library/books/:bookID/edit" element={<EditBook />} />
          </Route>
          <Route path="*" element={<NotFound />}/>
        </Routes>
      </Suspense>
    </AuthProvider>
  );
}

export default App;
