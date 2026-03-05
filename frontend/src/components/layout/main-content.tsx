interface MainContentProps {
  children: React.ReactNode;
}

export function MainContent({ children }: MainContentProps) {
  return (
    <main
      className="flex-1 bg-background"
      style={{ minWidth: "600px" }}
    >
      {children}
    </main>
  );
}
