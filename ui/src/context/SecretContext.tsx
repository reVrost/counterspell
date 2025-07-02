import { createContext, useContext, ReactNode } from "react";
import { useLocalStorage } from "@mantine/hooks";

type SecretContextType = {
  secret: string;
  setSecret: (secret: string) => void;
};

const SecretContext = createContext<SecretContextType | undefined>(undefined);

export const SecretProvider = ({ children }: { children: ReactNode }) => {
  const [secret, setSecret] = useLocalStorage<string>({
    key: "secret-token",
    defaultValue: "",
  });

  return (
    <SecretContext.Provider value={{ secret, setSecret }}>
      {children}
    </SecretContext.Provider>
  );
};

export const useSecret = () => {
  const context = useContext(SecretContext);
  if (context === undefined) {
    throw new Error("useSecret must be used within a SecretProvider");
  }
  return context;
};
