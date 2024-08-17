import { useEffect, useRef } from 'react';

const useScrollTransform = () => {
  const imageRef = useRef<HTMLImageElement>(null);

  useEffect(() => {
    const handleScroll = () => {
      const image = imageRef.current;
      if (!image) return;

      const rect = image.getBoundingClientRect();
      const scrollY = window.scrollY || document.documentElement.scrollTop;
      const maxScroll = window.innerHeight;
      const scrollRatio = Math.min(scrollY / maxScroll, 1);
      const scale = 1 - scrollRatio * 0.5;
      const opacity = 1 - scrollRatio * 4;
      const translateY = scrollRatio * -100;

      if (rect.top + rect.height > 0) {
        image.style.transform = `scale(${scale}) translateY(${translateY}px)`;
        image.style.opacity = `${opacity}`;
      }
    };

    window.addEventListener('scroll', handleScroll);

    return () => {
      window.removeEventListener('scroll', handleScroll);
    };
  }, [imageRef]);


  return imageRef;
};

export default useScrollTransform;
