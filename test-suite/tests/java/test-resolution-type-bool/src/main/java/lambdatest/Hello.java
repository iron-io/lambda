package lambdatest;

import com.amazonaws.services.lambda.runtime.Context;

import java.io.IOException;
import java.io.InputStream;
import java.io.OutputStream;
import java.util.ArrayList;
import java.util.List;
import java.util.Map;

public class Hello {
    public static void myHandler(boolean b) {
        System.out.println("Boolean " + b);
    }

    public static void myHandler(String s) {
        System.out.println("String " + s);
    }
}
