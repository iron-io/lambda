package lambdatest;

import com.amazonaws.services.lambda.runtime.Context;

import java.io.IOException;
import java.io.InputStream;
import java.io.OutputStream;
import java.util.ArrayList;
import java.util.List;
import java.util.Map;

public class Hello {
    public static class Atom {
      String symbol;
      String name;
      int atomicNumber;
      boolean foundOnEarth;
      List<Double> configuration;
      Map<String, Object> halfLife;

      public String getSymbol() {
        return symbol;
      }

      public String getName() {
        return name;
      }

      public int getAtomicNumber() {
        return atomicNumber;
      }

      public boolean getFoundOnEarth() {
        return foundOnEarth;
      }

      public List<Double> getConfiguration() {
        return configuration;
      }

      public Map<String, Object> getHalfLife() {
        return halfLife;
      }

      public void setSymbol(String s) {
        symbol = s;
      }

      public void setName(String n) {
        name = n;
      }

      public void setAtomicNumber(int a) {
        atomicNumber = a;
      }

      public void setFoundOnEarth(boolean f) {
        foundOnEarth = f;
      }

      public void setConfiguration(List<Double> l) {
        configuration = l;
      }

      public void setHalfLife(Map<String, Object> m) {
        halfLife = m;
      }

    }

    public static void myHandler(Atom a, Context c) {
        System.out.println("Symbol: " + a.symbol);
        System.out.println("Name: " + a.name);
        System.out.println("Atomic Number: " + a.atomicNumber);
        System.out.println("Found on Earth: " + a.foundOnEarth);
        System.out.println("Configuration: " + a.configuration);
        System.out.println("Half lives: " + a.halfLife);
    }
}
